# test
## 1111
### 22222
## b
## c
~~~c
#pragma once
#include <stdio.h>
#include <stdlib.h>
class class_c {
public:
  class_c() {}
  void run_class_c();
  int  run_class_1(int a,int c){
	return 0;  
  }
  void call_1(int a,int b) {}
  void call_2() { 
  	call_1(1,2);
  }  
}  

~~~
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