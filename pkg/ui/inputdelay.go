package mainui

import "time"

type inputdelay struct {
	// cb func(word string)
	//keymap  map[string]func()
	keyseq       string
	cmdlist      []cmditem
	main         *mainui
	delay_cmd    *cmditem
	delay_cmd_cb func()
}

type cmd_action int

const (
	cmd_action_none cmd_action = iota
	cmd_action_run
	cmd_action_delay
	cmd_action_buffer
)

func (input *inputdelay) get_cmd(key string) (cmd_action, []*cmditem) {
	var cmds = []*cmditem{}
	matched := 0
	same := 0
	input.delay_cmd = nil
	cmdlist := input.cmdlist
	for i, _ := range cmdlist {
		v := &cmdlist[i]
		if v.key.prefixmatched(key) {
			matched++
		}
		if v.key.string() == key {
			same++
			cmds = append(cmds, v)
		}
	}
	if len(cmds) > 0 {
		if matched > 1 {
			return cmd_action_delay, cmds
		} else {
			return cmd_action_run, cmds
		}
	}
	if matched > 1 {
		return cmd_action_buffer, cmds
	}
	return cmd_action_none, cmds
}
func (input *inputdelay) run_delay_cmd(cmd *cmditem) {
	input.delay_cmd = cmd
	go func() {
		timer := time.NewTimer(time.Millisecond * 200) // 两秒后触发
		<-timer.C
		defer timer.Stop()
		if input.main != nil {
			input.main.app.QueueUpdate(func() {
				if input.delay_cmd != nil {
					input.delay_cmd.cmd.handle()
					if input.delay_cmd_cb != nil {
						input.delay_cmd_cb()
					}
					input.main.app.ForceDraw()
				}
			})
		}
	}()
}

func (input *inputdelay) check(cmd string) cmd_action {
	action, cmds := input.get_cmd(cmd)
	switch action {
	case cmd_action_run:
		{
			var cmd = cmds[0]
			cmd.cmd.handle()
		}
	case cmd_action_delay:
		if len(cmds) > 0 {
			input.run_delay_cmd(cmds[0])
		}
	}
	return action
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

// func (input *inputdelay) rundelay(word string) {
// 	go func() {
// 		timer := time.NewTimer(time.Millisecond * 200) // 两秒后触发
// 		<-timer.C
// 		defer timer.Stop()
// 		input.main.app.QueueUpdate(func() {
// 			// input.cb(word)
// 			input.main.app.ForceDraw()
// 		})
// 	}()
// }
