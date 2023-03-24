package proof

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/polynetwork/poly/common"
	"github.com/qqtou/eos-go"
)

type MerkleTree struct {
	Root       *Node
	merkleRoot []byte
	Leafs      []*Node
}

type Node struct {
	Parent *Node
	Left   *Node
	Right  *Node
	leaf   bool
	dup    bool
	Hash   []byte // hash结果
	C      []byte // 原数据序列化的[]byte, 也可以后续优化
}

var (
	LeftSign  uint64 = 0xFFFFFFFFFFFFFF7F
	RightSign uint64 = 0x0000000000000080
	emptyRoot        = make([]byte, 32)
)

/*
计算哈希
*/
func CalculateHash(hash []byte) []byte {
	h := sha256.New()
	_, _ = h.Write(hash)
	return h.Sum(nil)
}

/*
计算节点哈希
*/
func CalculateNodeHash(left, right []byte) []byte {
	temp := append(left, right...)
	return CalculateHash(temp)
}

/*
原数据序列化
*/
func SerializationTrans(tr eos.TransactionReceipt) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tr.Status)
	binary.Write(buf, binary.LittleEndian, tr.CPUUsageMicroSeconds)
	binary.Write(buf, binary.LittleEndian, tr.NetUsageWords)

	binary.Write(buf, binary.LittleEndian, tr.Transaction.ID)
	binary.Write(buf, binary.LittleEndian, tr.Transaction.Packed)
	return buf.Bytes()
}

/*
左节点判断
*/
func JudgeLeft(left []byte) bool {
	return (left[0] & byte(RightSign)) == 0
}

/*
右节点判断
*/
func Judgeright(right []byte) bool {
	return (right[0] & byte(RightSign)) != 0
}

/*
左节点标记
*/
func SignToLeft(node []byte) []byte {
	left := node
	left[0] &= byte(LeftSign)

	return left
}

/*
右节点标记
*/
func SignToRight(node []byte) []byte {
	right := node
	right[0] |= byte(RightSign)
	return right
}

/*
计算所有叶子节点的Hash再递归调用buildIntermediate 构建整棵树
*/
func buildWithLeaf(cs []eos.TransactionReceipt, t *MerkleTree) (*Node, []*Node, error) {
	if len(cs) == 0 {
		return nil, nil, nil
	}
	var leafs []*Node
	var hashed []byte
	for _, c := range cs {
		temp := SerializationTrans(c)
		hashed = CalculateHash(temp)

		leafs = append(leafs, &Node{
			Hash: hashed,
			C:    temp,
			leaf: true,
			dup:  false,
		})
	}
	// 补足偶数
	if len(leafs)%2 == 1 {
		var reHash = make([]byte, 32)
		copy(reHash, leafs[len(leafs)-1].Hash)

		duplicate := &Node{
			Hash: reHash,
			C:    leafs[len(leafs)-1].C,
			leaf: true,
			dup:  true,
		}
		leafs = append(leafs, duplicate)
	}

	root, err := buildIntermeddiate(leafs, t)
	if err != nil {
		return nil, nil, err
	}
	return root, leafs, nil
}

/*
对于给定叶子节点列表，构建树的中间层和根，并返回树的根节点
*/
func buildIntermeddiate(nl []*Node, t *MerkleTree) (*Node, error) {
	var nodes []*Node
	//补足
	if len(nl)%2 == 1 {
		var reHash = make([]byte, 32)
		copy(reHash, nl[len(nl)-1].Hash)

		duplicate := &Node{
			Hash: reHash,
			C:    nl[len(nl)-1].C,
			leaf: false,
			dup:  true,
		}
		nl = append(nl, duplicate)
	}

	for i := 0; i < len(nl)/2; i++ {
		var left, right int = i * 2, i*2 + 1
		if i+1 == len(nl) {
			right = i
		}

		hashed := CalculateNodeHash(SignToLeft(nl[left].Hash), SignToRight(nl[right].Hash))

		n := &Node{
			Left:  nl[left],
			Right: nl[right],
			Hash:  hashed,
			leaf:  false,
			dup:   false,
		}

		nodes = append(nodes, n)
		nl[left].Parent = n
		nl[right].Parent = n
	}

	if len(nl) == 2 {
		return nodes[0], nil
	}
	return buildIntermeddiate(nodes, t)
}

/*
获取默克尔根
*/
func (m *MerkleTree) GetMerkleRoot() []byte {
	return m.merkleRoot
}

/*
构造默克尔树
*/
func NewTree(cs []eos.TransactionReceipt) (*MerkleTree, error) {
	t := &MerkleTree{}
	root, leafs, err := buildWithLeaf(cs, t)
	if err != nil {
		return nil, err
	}
	if leafs != nil {
		t.merkleRoot = root.Hash
	} else {
		t.merkleRoot = emptyRoot
	}
	t.Root = root
	t.Leafs = leafs

	return t, nil
}

/*
获取默克尔证明的路径
*/
func (m *MerkleTree) GetMerklePath(c []byte) ([][]byte, []byte, error) {
	for i, current := range m.Leafs {
		if bytes.Equal(c, current.C) && current.dup == false {
			currentParent := current.Parent
			var merklePath [][]byte
			for currentParent != nil {
				if bytes.Equal(currentParent.Left.Hash, current.Hash) {
					merklePath = append(merklePath, currentParent.Right.Hash)
				} else {
					merklePath = append(merklePath, currentParent.Left.Hash)
				}
				current = currentParent
				currentParent = currentParent.Parent
			}
			return merklePath, m.Leafs[i].Hash, nil
		}

	}
	return nil, nil, nil
}

/*
自底向上层次遍历输出默克尔树的每一个节点
*/
func (m *MerkleTree) ToString() error {
	res := make([][]byte, 0)
	leafs := m.Leafs
	if leafs == nil {
		return fmt.Errorf("Merkle Tree Leafs is nil")
	}
	queue := make([]*Node, 0)
	queue = append(queue, leafs...)
	for len(queue) > 0 {
		if len(queue) == 1 {
			fmt.Printf("the queue len is 1 and the node is root, hash is%v\n", queue[0].Hash)
			res = append(res, queue[0].Hash)
			break
		}
		if len(queue)%2 != 0 {
			queue = append(queue, queue[len(queue)-1])
		}
		l := len(queue)
		fmt.Printf("the len of queue is%d\n", l)
		for l >= 2 {
			left := queue[0]
			queue = queue[1:]
			right := queue[0]
			queue = queue[1:]
			res = append(res, left.Hash, right.Hash)
			if left.Parent != nil && right.Parent != nil {
				queue = append(queue, left.Parent)
			}
			l -= 2
		}

	}
	for i := 0; i < len(res); i++ {
		fmt.Printf("the %d node Hash is %v\n", i, res[i])
	}
	return nil
}

/*
验证默克尔证明
*/

func VerifyLeaf(path [][]byte, leaf []byte) []byte {

	tempLeaf := make([]byte, 32)
	tempPath := make([]byte, 32)
	copy(tempLeaf, leaf)
	for _, pa := range path {
		copy(tempPath, pa)
		if JudgeLeft(tempPath) {
			tempLeaf = SignToRight(tempLeaf)
			tempLeaf = CalculateNodeHash(tempPath, tempLeaf)
		} else {
			tempLeaf = SignToLeft(tempLeaf)
			tempLeaf = CalculateNodeHash(tempLeaf, tempPath)
		}
	}
	return tempLeaf
}

// func VerifyLeaf(path [][]byte, leaf []byte) []byte {

// 	if len(path) < 1 {

// 		return leaf
// 	}

// 	pa := path[0]
// 	path = path[1:]

// 	var temp []byte
// 	if JudgeLeft(pa) {
// 		releaf := SignToRight(leaf)

// 		temp = CalculateNodeHash(pa, releaf)

// 	} else {
// 		releaf := SignToLeft(leaf)

// 		temp = CalculateNodeHash(releaf, pa)

// 	}

// 	return VerifyLeaf(path, temp)
// }

type EOSProof struct {
	path [][]byte
	leaf []byte
}

func (this *EOSProof) Serialization(sink *common.ZeroCopySink) {
	sink.WriteBytes(this.leaf)
	// fmt.Printf("Write leaf len is:%v\n", len(this.leaf))
	for i := 0; i < len(this.path); i++ {
		sink.WriteBytes(this.path[i])
		// fmt.Printf("Write path %d is:%v\n", i, len(this.path[i]))
	}
}

func (this *EOSProof) Deserialization(source *common.ZeroCopySource) error {
	n := source.Len()
	// fmt.Printf("source len is:%d", n)
	if (n % 32) != 0 {
		return fmt.Errorf("Deserialization error : len is illegal")
	}
	var path [][]byte
	// var leaf []byte
	leaf, eof := source.NextBytes(32)
	if eof {
		return fmt.Errorf("Waiting deserialize leaf error")
	}

	for i := 0; i < int(n/32)-1; i++ {
		pa, eof := source.NextBytes(32)
		fmt.Printf("the source the %d pa len is: %d\n", i+1, source.Len())
		if eof {
			return fmt.Errorf("Waiting deserialize %d path error", i)
		}
		path = append(path, pa)
	}

	this.leaf = leaf
	this.path = path
	return nil
}
