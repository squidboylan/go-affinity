package mesh

import (
	_ "github.com/project-receptor/receptor/pkg/backends"
	_ "github.com/project-receptor/receptor/pkg/netceptor"
	_ "io"
	_ "net"
	_ "os"
	_ "sync"
	_ "time"
	"testing"
)

func TestNode(t *testing.T) {
	testTable := []struct {
		filename string
	} {
		{"test-mesh.yaml"},
		{"random-mesh.yaml"},
		{"flat-mesh.yaml"},
	}
	for _, data := range testTable {
		t.Run(data.filename, func(t *testing.T) {
			mesh := NewMeshFromFile(data.filename)
			t.Log(mesh.nodes)
			t.Log(mesh.connections)
		})
	}
}
