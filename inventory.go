package main

func InventoryRoutes(s *Server) {
	s.router.GET("/inventory", noOp(s))
}
