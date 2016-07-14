package context

import (
	"net/http"
)

//type context struct {
//	data map[*http.Request]map[interface{}]interface{}
//}

//var c context = context{make(map[*http.Request]map[interface{}]interface{})}

//func (con *context) Set(r *http.Request, key, value interface{}) {
//	_, ok := con.data[r]
//	if !ok {
//		con.data[r] = make(map[interface{}]interface{})
//	}
//	con.data[r][key] = value
//
//}
//func (con context) Get(r *http.Request, key interface{}) interface{} {
//	return con.data[r][key]
//}

var context = make(map[*http.Request]map[interface{}]interface{})

func Set(r *http.Request, key, value interface{}) {
	_, ok := context[r]
	if !ok {
		context[r] = make(map[interface{}]interface{})
	}
	context[r][key] = value

}
func Get(r *http.Request, key interface{}) interface{} {
	return context[r][key]
}
