package mainui

func RunGrep(grep *gorep, fpath string, out chan<- grepInfo) bool {
	return PosixRunGrep(grep, fpath, out)
}
