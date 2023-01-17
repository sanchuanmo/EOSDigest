package proof

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

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

func CalculateHash(hash []byte) []byte {
	h := sha256.New()
	_, _ = h.Write(hash)
	return h.Sum(nil)
}

func CalculateNodeHash(left, right []byte) []byte {
	temp := append(left, right...)
	return CalculateHash(temp)
}

func SerializationTrans(tr eos.TransactionReceipt) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, tr.Status)
	binary.Write(buf, binary.LittleEndian, tr.CPUUsageMicroSeconds)
	binary.Write(buf, binary.LittleEndian, tr.NetUsageWords)
	// temp, _ := tr.Transaction.MarshalJSON()
	// binary.Write(buf, binary.LittleEndian, temp)
	binary.Write(buf, binary.LittleEndian, tr.Transaction.ID)
	binary.Write(buf, binary.LittleEndian, tr.Transaction.Packed)
	return buf.Bytes()
}

// func DeserializationTrans(trByte []byte) eos.TransactionReceipt {

// }

func JudgeLeft(left []byte) bool {
	return (left[0] & byte(RightSign)) == 0
}
func Judgeright(right []byte) bool {
	return (right[0] & byte(RightSign)) != 0
}

func SignToLeft(node []byte) []byte {
	left := node
	left[0] &= byte(LeftSign)

	return left
}
func SignToRight(node []byte) []byte {
	right := node
	right[0] |= byte(RightSign)
	return right
}

// 计算所有叶子节点的Hash再递归调用buildIntermediate 构建整棵树
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
		// fmt.Printf("reHash: %v\n", reHash)
		duplicate := &Node{
			Hash: reHash,
			C:    leafs[len(leafs)-1].C,
			leaf: true,
			dup:  true,
		}
		leafs = append(leafs, duplicate)
	}

	// for i := 0; i < len(leafs); i++ {
	// 	fmt.Printf("the leaf[%d] hashed is:%v\n", i, leafs[i].Hash)
	// }
	// fmt.Printf("\n")

	root, err := buildIntermeddiate(leafs, t)
	if err != nil {
		return nil, nil, err
	}
	return root, leafs, nil
}

// 对于给定叶子节点列表，构建树的中间层和根，并返回树的根节点
func buildIntermeddiate(nl []*Node, t *MerkleTree) (*Node, error) {
	var nodes []*Node
	//补足
	if len(nl)%2 == 1 {
		var reHash = make([]byte, 32)
		copy(reHash, nl[len(nl)-1].Hash)
		// fmt.Printf("reHash: %v\n", reHash)
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
		// fmt.Printf("nl[left].Hash: %v\n", nl[left].Hash)
		// fmt.Printf("nl[right].Hash: %v\n", nl[right].Hash)
		hashed := CalculateNodeHash(SignToLeft(nl[left].Hash), SignToRight(nl[right].Hash))
		// fmt.Printf("nl[left].Hash: %v\n", nl[left].Hash)
		// fmt.Printf("nl[right].Hash: %v\n", nl[right].Hash)
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

	// for i := 0; i < len(nodes); i++ {
	// 	fmt.Printf("the node hashed is:%v\n", nodes[i].Hash)
	// }
	// fmt.Printf("\n")
	if len(nl) == 2 {
		return nodes[0], nil
	}
	return buildIntermeddiate(nodes, t)
}

func (m *MerkleTree) GetMerkleRoot() []byte {
	return m.merkleRoot
}

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
已优化，无需index
*/
// func (m *MerkleTree) GetMerklePath(c []byte) ([][]byte, []uint64, error) {
func (m *MerkleTree) GetMerklePath(c []byte) ([][]byte, []byte, error) {
	for i, current := range m.Leafs {
		// ok :=
		// fmt.Printf("Get the %d leaf in tree, leaf hash is: %v", i, current.Hash)
		if bytes.Equal(c, current.C) && current.dup == false {
			fmt.Printf("Get the node in tree, index is %d\n,the goal node: %v\n match node: %v\n", i, CalculateHash(c), current.Hash)
			currentParent := current.Parent
			var merklePath [][]byte
			// var index []uint64
			for currentParent != nil {
				fmt.Printf("currentParent is :%v", currentParent.Hash)
				if bytes.Equal(currentParent.Left.Hash, current.Hash) {
					merklePath = append(merklePath, currentParent.Right.Hash)
					fmt.Printf("Append the right Hash,%v", currentParent.Right.Hash)
					// index = append(index, 1) // right leaf
				} else {
					merklePath = append(merklePath, currentParent.Left.Hash)
					// index = append(index, 0) // left leaf
					fmt.Printf("Append the left Hash,%v", currentParent.Left.Hash)
				}
				current = currentParent
				currentParent = currentParent.Parent
			}
			return merklePath, m.Leafs[i].Hash, nil
		}

	}
	return nil, nil, nil
}

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

func VerifyLeaf(path [][]byte, leaf []byte) []byte {

	if len(path) < 1 {
		fmt.Printf("root is :%v\n", leaf)
		return leaf
	}

	pa := path[0]
	path = path[1:]

	var temp []byte
	if JudgeLeft(pa) {
		releaf := SignToRight(leaf)
		fmt.Printf("Signed right leaf:%v\n, left node:%v\n", releaf, pa)
		temp = CalculateNodeHash(pa, releaf)
		fmt.Printf("the next Node is:%v\n", temp)
	} else {
		releaf := SignToLeft(leaf)
		fmt.Printf("Signed left leaf:%v\n, right node:%v\n", leaf, pa)
		temp = CalculateNodeHash(releaf, pa)
		fmt.Printf("the next Node is:%v\n", temp)
	}

	return VerifyLeaf(path, temp)
}
