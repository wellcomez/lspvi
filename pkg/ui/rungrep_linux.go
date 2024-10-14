package mainui

func RunGrep(grep *gorep, fpath string, out chan<- GrepInfo) {
	PosixRunGrep(grep, fpath, out)
}
