package cmd

import (
	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/sirupsen/logrus"
	"github.com/xlab/treeprint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"strings"
)

var (
	nilID = "00000000-0000-0000-0000-000000000000"
)

func createConnection(port string) (*grpc.ClientConn, error) {
	return grpc.Dial(port, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func sanitizeID(id string) string {
	id = strings.ToLower(id)

	if len(id) <= 36 {
		idLen := len(id)
		return nilID[:36-idLen] + id
	}

	if len(id) > 36 {
		return id[:36]
	}

	return id
}

func printBlock(block *v1.Block) {
	var build func(block *v1.Block, tree treeprint.Tree)
	build = func(block *v1.Block, tree treeprint.Tree) {
		if block.Children != nil && len(block.Children) != 0 {
			children := tree.AddBranch("children")
			for _, child := range block.Children {
				children.AddNode(child.BlockId)
				build(child, children)
			}
		}

		if block.Linked != nil && len(block.Linked) != 0 {
			linked := tree.AddBranch("linked")
			for _, link := range block.Linked {
				linked.AddNode(link.BlockId)
				build(link, linked)
			}
		}
	}

	tree := treeprint.New()
	tree.AddNode(block.BlockId)
	build(block, tree)

	logrus.Info(tree.String())
}
