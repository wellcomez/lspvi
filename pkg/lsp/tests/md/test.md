# test
## 1111
### 22222
## b
## c
~~~go
    func (t *Tree) load_ts_lang(cb chan<- bool) error {}
	for i := range tree_sitter_lang_map {
		v := tree_sitter_lang_map[i]
		if ts_name := v.get_ts_name(t.filename.Path()); len(ts_name) > 0 {
			v.load_scm()
			t.tsdef = v
			// t.Loadfile(v.tslang, cb)
			t.tsdef.intiqueue.Run(ts_init_call{t, cb, ts_load_call, nil})
			return nil
		}
	}
~~~