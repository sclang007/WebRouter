package main

import (
    "bytes"
    "fmt"
	"io"
	"os"
    "log"
	"net"
	"strings"
	"regexp"
	"io/ioutil"
	"encoding/json"
)
type Rule struct{
	Domain string `json:Domain`
	Address string `json:Address`
}
type Config struct{
	MainPort string `json:MainPort`
	Rules []Rule `json:Rules`
}
var myConfig Config
func InitConfig(){
	var Data,err = ioutil.ReadFile("config.json")
	if err != nil{
		log.Println("Read Config File Error！")
		os.Exit(0)
		return
	}
	err = json.Unmarshal(Data,&myConfig)
	if err != nil{
		log.Println("Read Config JSON Error！Please Check!")
		os.Exit(0)
		return
	}
	fmt.Println("Main Port:"+myConfig.MainPort)
	for i:=0;i<len(myConfig.Rules);i++{
		fmt.Println("Domain: "+myConfig.Rules[i].Domain+" <----> "+myConfig.Rules[i].Address)
	}
}
func handleClientRequest(client net.Conn) {
    if client == nil {
        return
    }
    defer client.Close()
    var b [1024]byte
    n, err := client.Read(b[:])
    if err != nil {
        log.Println(err)
        return
    }
    var method, url, HTTPv, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s%s", &method, &url, &HTTPv)
	address = GetAddress(string(b[:]))
	if address == "nil"{
		fmt.Println("Unknow Domain")
		return
	}
    server, err := net.Dial("tcp", address)
    if err != nil {
        log.Println(err)
        return
    }
    if method == "CONNECT" {
        fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
    } else {
        server.Write(b[:n])
    }
    go io.Copy(server, client)
	io.Copy(client, server)
}
func GetAddress(HTTPData string) string{
	Lines := strings.Split(HTTPData,"\r\n")
	for i:=0;i<len(Lines);i++{
		line := Lines[i]
		temp := strings.Split(line,": ")
		if strings.Compare(temp[0],"Host")==0{
			for i:=0;i<len(myConfig.Rules);i++{
				if strings.Compare(myConfig.Rules[i].Domain,temp[1])==0{
					return myConfig.Rules[i].Address
				}
			}
		}
	}
	return "nil"
}
func compressStr(str string) string {
    if str == "" {
        return ""
    }
    reg := regexp.MustCompile("\\s+")
    return reg.ReplaceAllString(str, "")
}
func main() {
	log.SetFlags(log.LstdFlags|log.Lshortfile)
	InitConfig()
    l, err := net.Listen("tcp", ":"+myConfig.MainPort)
    if err != nil {
        log.Panic(err)
    }
    for {
        client, err := l.Accept()
        if err != nil {
            log.Panic(err)
        }
        go handleClientRequest(client)
    }
}