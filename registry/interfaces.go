package registry

type Register interface {
	NodeExists(string) bool
	AddNode(string) error
	RemoveNode(string) error
	// todo:
	// GetNodeByName
	// GetNodeByIpaddr
	// GetNodeByNet (на входе network obj.)

	// GetNodesGraph
	//  -относительно себя
	//  -относительно центра
	//  -относительно рэка
	// GetRackNodes
	// GetRackNodesForRack
	// GetShortestPath
	// GetQuickestPath
}
