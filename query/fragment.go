package query

type fragment interface {
	compile(c *Compiler)
}
