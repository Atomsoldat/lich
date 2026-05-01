package lich


// TODO: if we add fields for "processing", "completed", "success" and so on
// we can parallelise the templating
type UnitOfWork struct {
	Name          string
	Origin        string
	Kustomization string
	Destination   string
}
