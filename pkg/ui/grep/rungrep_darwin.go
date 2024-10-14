package grep

func RunGrep(grep *Gorep, fpath string, out chan<- GrepInfo) {
	PosixRunGrep(grep, fpath, out)
}
