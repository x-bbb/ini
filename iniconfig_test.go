package iniconfig

import (
	"testing"
)

type Mysql struct {
	Username string `ini:"username"`
	Password string `ini:"password"`
	Port     int    `ini:"port"`
	Host     string `ini:"host"`
	Database string `ini:"database"`
}

type Config struct {
	Server Server `ini:"server"`
	Mysql  Mysql  `ini:"mysql"`
}

type Server struct {
	IP   string `ini:"ip"`
	Port int    `ini:"port"`
}

func TestIniConfig(t *testing.T) {
	// data, err := ioutil.ReadFile("test.ini")
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
	var conf Config

	// err = UnMarshal(data, &conf)
	// if err != nil {
	// 	t.Fatal(err)
	// 	return
	// }
	//
	// err = MarshalFile("d:/test.conf", conf)
	//
	// // data, err = Marshal(conf)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// //fmt.Println(string(data))

	err := UnMarshalFile("d:/test.conf", &conf)

	if err != nil {
		t.Fatal(err)
	}

	t.Logf("unmarshal success, conf:%#v\n", conf)

}
