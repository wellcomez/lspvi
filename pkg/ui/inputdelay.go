package mainui

import "time"
type inputdelay struct {
	cb func(word string)
	//keymap  map[string]func()
	keyseq  string
	cmdlist []cmditem
	main* mainui
}

func (input *inputdelay) command_matched(key string) int {
	matched := 0
	/*if input.keymap != nil {
		for k := range input.keymap {
			if strings.HasPrefix(k, key) {
				matched++
			}
		}
	} */
	cmdlist := input.cmdlist
	for _, v := range cmdlist {
		if v.key.matched(key) {
			matched++
		}
	}
	return matched
}
func (input *inputdelay) check(cmd string) (bool, bool) {
	matched := input.command_matched(cmd)
	if matched == 1 {
		if input.run(cmd)==false{
			return false, false 
		}
	}
	return matched > 0, matched == 1
}
func (input *inputdelay) run(cmd string) bool {
	/*if input.keymap != nil {
		if cb, ok := input.keymap[cmd]; ok {
			cb()
			input.keyseq = ""
			return ok
		}
	}*/
	if input.cmdlist != nil {
		for _, v := range input.cmdlist {
			if v.key.string() == cmd {
				v.cmd.handle()
				return true
			}
		}
	}
	return false
}
func (input *inputdelay) rundelay(word string) {
	go func() {
		timer := time.NewTimer(time.Millisecond * 200) // 两秒后触发
		<-timer.C
		defer timer.Stop()
		input.main.app.QueueUpdate(func() {
			input.cb(word)
			input.main.app.ForceDraw()
		})
	}()
}
