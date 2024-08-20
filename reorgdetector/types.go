package reorgdetector

import (
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type BlockOrigin string

const (
	OriginSubscription BlockOrigin = "Subscription"
	OriginGetParent    BlockOrigin = "GetParent"
)

// Reorg is the analysis Summary of a specific reorg
type Reorg struct {
	IsFinished bool
	SeenLive   bool

	StartBlockHeight uint64 // first block in a reorg (block number after common parent)
	EndBlockHeight   uint64 // last block in a reorg

	Chains map[common.Hash][]*Block

	Depth int

	BlocksInvolved    map[common.Hash]*Block
	MainChainHash     common.Hash
	MainChainBlocks   map[common.Hash]*Block
	NumReplacedBlocks int

	CommonParent         *Block
	FirstBlockAfterReorg *Block
}

func NewReorg(parentNode *TreeNode) (*Reorg, error) {
	if len(parentNode.Children) < 2 {
		return nil, fmt.Errorf("cannot create reorg because parent node with < 2 children")
	}

	reorg := Reorg{
		CommonParent:     parentNode.Block,
		StartBlockHeight: parentNode.Block.Number() + 1,

		Chains:          make(map[common.Hash][]*Block),
		BlocksInvolved:  make(map[common.Hash]*Block),
		MainChainBlocks: make(map[common.Hash]*Block),

		SeenLive: true, // will be set to false if any of the added blocks was received via uncle-info
	}

	// Build the individual chains until the enc, by iterating over children recursively
	for _, chainRootNode := range parentNode.Children {
		chain := make([]*Block, 0)

		var addChildToChainRecursive func(node *TreeNode)
		addChildToChainRecursive = func(node *TreeNode) {
			chain = append(chain, node.Block)

			for _, childNode := range node.Children {
				addChildToChainRecursive(childNode)
			}
		}
		addChildToChainRecursive(chainRootNode)
		reorg.Chains[chainRootNode.Block.Header.Hash()] = chain
	}

	// Find depth of chains
	chainLengths := []int{}
	for _, chain := range reorg.Chains {
		chainLengths = append(chainLengths, len(chain))
	}
	sort.Sort(sort.Reverse(sort.IntSlice(chainLengths)))

	// Depth is number of blocks in second chain
	reorg.Depth = chainLengths[1]

	// Truncate the longest chain to the second, which is when the reorg actually stopped
	for key, chain := range reorg.Chains {
		if len(chain) > reorg.Depth {
			reorg.FirstBlockAfterReorg = chain[reorg.Depth] // first block that will be truncated
			reorg.Chains[key] = chain[:reorg.Depth]
			reorg.MainChainHash = key
		}
	}

	// If two chains with same height, then the reorg isn't yet finalized
	if chainLengths[0] == chainLengths[1] {
		reorg.IsFinished = false
	} else {
		reorg.IsFinished = true
	}

	// Build final list of involved blocks, and get end blockheight
	for chainHash, chain := range reorg.Chains {
		for _, block := range chain {
			reorg.BlocksInvolved[block.Header.Hash()] = block

			if block.Origin != OriginSubscription && block.Origin != OriginGetParent {
				reorg.SeenLive = false
			}

			if chainHash == reorg.MainChainHash {
				reorg.MainChainBlocks[block.Header.Hash()] = block
				reorg.EndBlockHeight = block.Number()
			}
		}
	}

	reorg.NumReplacedBlocks = len(reorg.BlocksInvolved) - reorg.Depth

	return &reorg, nil
}

func (r *Reorg) Id() string {
	id := fmt.Sprintf("%d_%d_d%d_b%d", r.StartBlockHeight, r.EndBlockHeight, r.Depth, len(r.BlocksInvolved))
	if r.SeenLive {
		id += "_l"
	}
	return id
}

func (r *Reorg) String() string {
	return fmt.Sprintf("Reorg %s: live=%-5v chains=%d, depth=%d, replaced=%d", r.Id(), r.SeenLive, len(r.Chains), r.Depth, r.NumReplacedBlocks)
}

func (r *Reorg) MermaidSyntax() string {
	ret := "stateDiagram-v2\n"

	for _, block := range r.BlocksInvolved {
		ret += fmt.Sprintf("    %s --> %s\n", block.Header.ParentHash, block.Header.Hash())
	}

	// Add first block after reorg
	ret += fmt.Sprintf("    %s --> %s", r.FirstBlockAfterReorg.Header.ParentHash, r.FirstBlockAfterReorg.Header.Hash())
	return ret
}

type TreeNode struct {
	Block    *Block
	Parent   *TreeNode
	Children []*TreeNode

	IsFirst     bool
	IsMainChain bool
}

func NewTreeNode(block *Block, parent *TreeNode) *TreeNode {
	return &TreeNode{
		Block:    block,
		Parent:   parent,
		Children: []*TreeNode{},
		IsFirst:  parent == nil,
	}
}

func (tn *TreeNode) String() string {
	return fmt.Sprintf("TreeNode %d %s main=%5v \t first=%5v, %d children", tn.Block.Number(), tn.Block.Header.Hash(), tn.IsMainChain, tn.IsFirst, len(tn.Children))
}

func (tn *TreeNode) AddChild(node *TreeNode) {
	tn.Children = append(tn.Children, node)
}

// TreeAnalysis takes in a BlockTree and collects information about reorgs
type TreeAnalysis struct {
	Tree *BlockTree

	StartBlockHeight uint64 // first block number with siblings
	EndBlockHeight   uint64
	IsSplitOngoing   bool

	NumBlocks          int
	NumBlocksMainChain int

	Reorgs map[string]*Reorg
}

func NewTreeAnalysis(t *BlockTree) (*TreeAnalysis, error) {
	analysis := TreeAnalysis{
		Tree:   t,
		Reorgs: make(map[string]*Reorg),
	}

	if t.FirstNode == nil { // empty analysis for empty tree
		return &analysis, nil
	}

	analysis.StartBlockHeight = t.FirstNode.Block.Number()
	analysis.EndBlockHeight = t.LatestNodes[0].Block.Number()

	if len(t.LatestNodes) > 1 {
		analysis.IsSplitOngoing = true
	}

	analysis.NumBlocks = len(t.NodeByHash)
	analysis.NumBlocksMainChain = len(t.MainChainNodeByHash)

	// Find reorgs
	for _, node := range t.NodeByHash {
		if len(node.Children) > 1 {
			reorg, err := NewReorg(node)
			if err != nil {
				return nil, err
			}
			analysis.Reorgs[reorg.Id()] = reorg
		}
	}

	return &analysis, nil
}

func (a *TreeAnalysis) Print() {
	fmt.Printf("TreeAnalysis %d - %d, nodes: %d, mainchain: %d, reorgs: %d\n", a.StartBlockHeight, a.EndBlockHeight, a.NumBlocks, a.NumBlocksMainChain, len(a.Reorgs))
	if a.IsSplitOngoing {
		fmt.Println("- split ongoing")
	}

	for _, reorg := range a.Reorgs {
		fmt.Println("")
		fmt.Println(reorg.String())
		fmt.Printf("- common parent: %d %s, first block after: %d %s\n", reorg.CommonParent.Header.Number.Uint64(), reorg.CommonParent.Header.Hash(), reorg.FirstBlockAfterReorg.Header.Number.Uint64(), reorg.FirstBlockAfterReorg.Header.Hash())

		for chainKey, chain := range reorg.Chains {
			if chainKey == reorg.MainChainHash {
				fmt.Printf("- mainchain l=%d: ", len(chain))
			} else {
				fmt.Printf("- sidechain l=%d: ", len(chain))
			}
			for _, block := range chain {
				fmt.Printf("%s ", block.Header.Hash())
			}
			fmt.Print("\n")
		}
	}
}

// BlockTree is the tree of blocks, used to traverse up (from children to parents) and down (from parents to children).
// Reorgs start on each node with more than one child.
type BlockTree struct {
	FirstNode           *TreeNode
	LatestNodes         []*TreeNode // Nodes at latest blockheight (can be more than 1)
	NodeByHash          map[common.Hash]*TreeNode
	MainChainNodeByHash map[common.Hash]*TreeNode
}

func NewBlockTree() *BlockTree {
	return &BlockTree{
		LatestNodes:         []*TreeNode{},
		NodeByHash:          make(map[common.Hash]*TreeNode),
		MainChainNodeByHash: make(map[common.Hash]*TreeNode),
	}
}

func (t *BlockTree) AddBlock(block *Block) error {
	// First block is a special case
	if t.FirstNode == nil {
		node := NewTreeNode(block, nil)
		t.FirstNode = node
		t.LatestNodes = []*TreeNode{node}
		t.NodeByHash[block.Header.Hash()] = node
		return nil
	}

	// All other blocks are inserted as child of it's parent parent
	parent, parentFound := t.NodeByHash[block.Header.ParentHash]
	if !parentFound {
		err := fmt.Errorf("error in BlockTree.AddBlock(): parent not found. block: %d %s, parent: %s", block.Number(), block.Header.Hash(), block.Header.ParentHash)
		return err
	}

	node := NewTreeNode(block, parent)
	t.NodeByHash[block.Header.Hash()] = node
	parent.AddChild(node)

	// Remember nodes at latest block height
	if len(t.LatestNodes) == 0 {
		t.LatestNodes = []*TreeNode{node}
	} else {
		if block.Number() == t.LatestNodes[0].Block.Number() { // add to list of latest nodes!
			t.LatestNodes = append(t.LatestNodes, node)
		} else if block.Number() > t.LatestNodes[0].Block.Number() { // replace
			t.LatestNodes = []*TreeNode{node}
		}
	}

	// Mark main-chain nodes as such. Step 1: reset all nodes to non-main-chain
	t.MainChainNodeByHash = make(map[common.Hash]*TreeNode)
	for _, n := range t.NodeByHash {
		n.IsMainChain = false
	}

	// Step 2: Traverse backwards and mark main chain. If there's more than 1 nodes at latest height, then we don't yet know which chain will be the main-chain
	if len(t.LatestNodes) == 1 {
		var TraverseMainChainFromLatestToEarliest func(node *TreeNode)
		TraverseMainChainFromLatestToEarliest = func(node *TreeNode) {
			if node == nil {
				return
			}
			node.IsMainChain = true
			t.MainChainNodeByHash[node.Block.Header.Hash()] = node
			TraverseMainChainFromLatestToEarliest(node.Parent)
		}
		TraverseMainChainFromLatestToEarliest(t.LatestNodes[0])
	}

	return nil
}

func (t *BlockTree) Print() {
	fmt.Printf("BlockTree: nodes=%d\n", len(t.NodeByHash))

	if t.FirstNode == nil {
		return
	}

	// Print tree by traversing from parent to all children
	PrintNodeAndChildren(t.FirstNode, 1)

	// Print latest nodes
	fmt.Printf("Latest nodes:\n")
	for _, n := range t.LatestNodes {
		fmt.Println("-", n.String())
	}
}

// Block is a geth Block and information about where it came from
type Block struct {
	Header *types.Header
	Origin BlockOrigin
}

func NewBlock(header *types.Header, origin BlockOrigin) *Block {
	return &Block{
		Header: header,
		Origin: origin,
	}
}

func (b *Block) Number() uint64 {
	return b.Header.Number.Uint64()
}

func PrintNodeAndChildren(node *TreeNode, depth int) {
	indent := "-"
	fmt.Printf("%s %s\n", indent, node.String())
	for _, childNode := range node.Children {
		PrintNodeAndChildren(childNode, depth+1)
	}
}
