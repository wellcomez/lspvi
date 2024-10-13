package mainui

func RunGrep(grep *gorep, fpath string, out chan<- GrepInfo) bool {
	return PosixRunGrep(grep, fpath, out)
}
