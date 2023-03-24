package proof

import (
	"encoding/hex"
	"fmt"
	"testing"
)

var (
	transactions = []string{"f7d465c918d185499dd09a7bbc9f46abe95063adbac3475771db62b0c081f49b", "912c5902895ccefa845ef4b6f08009c4f990f35afb45030a75833d027990b7ba",
		"c4c49479d4c1ccf3756b9021f6ad19049ff327ff663040a376dc4965f4caad1e", "65292ac69e8dac71ef2751c70b2659d0030d5b7b536f689840dd2c2d9e3bd08d"}
	chainID = "bf2344f6220947e518d8c69c39789b13525328d2f584aa10a4fb1d478869d523"
)

func NewTree2(cs []string) (*MerkleTree, error) {
	t := &MerkleTree{}
	root, leafs, err := buildWithLeaf2(cs, t)
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

func buildWithLeaf2(cs []string, t *MerkleTree) (*Node, []*Node, error) {
	if len(cs) == 0 {
		return nil, nil, nil
	}
	chainIDBytes, _ := hex.DecodeString(chainID)
	var leafs []*Node
	var hashed []byte
	for i, c := range cs {
		temp, err := hex.DecodeString(c)
		if err != nil {
			return nil, nil, fmt.Errorf("叶子节点%d转字节失败", i)
		}
		hashed = CalculateNodeHash(SignToLeft(chainIDBytes), SignToRight(temp))

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

func TestCalculateRoot(t *testing.T) {
	var leaf [][]byte
	for _, i := range transactions {
		temp, err := hex.DecodeString(i)
		if err != nil {
			fmt.Printf("转格式Error:%s\n", err)
		}
		leaf = append(leaf, temp)
	}
	node1 := CalculateNodeHash(leaf[0], leaf[1])
	node2 := CalculateNodeHash(leaf[2], leaf[3])
	root := CalculateNodeHash(node1, node2)
	fmt.Printf("cal root :%v", hex.EncodeToString(root))

}

func TestCalRoot(t *testing.T) {

}
