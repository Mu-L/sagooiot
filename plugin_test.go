package main

import (
	"fmt"
	"github.com/gogf/gf/v2/util/gconv"
	"net"
	"sagooiot/pkg/plugins"
	"sagooiot/pkg/plugins/model"
	"sagooiot/pkg/plugins/module"
	"testing"
)

func TestManagerInit(t *testing.T) {
	manager := plugins.NewManager("protocol", "protocol-*", "./plugins/built", &module.ProtocolPlugin{})
	defer manager.Dispose()
	err := manager.Init()
	if err != nil {
		t.Fatal(err.Error())
	}
	err = manager.Launch()
	for id, info := range manager.Plugins {
		t.Log(id)
		t.Log(info.Path)
		t.Log(info.Client)
	}

	t.Log(manager)

}

// 测试获取插件信息
func TestProtocolInfo(t *testing.T) {
	p, err := plugins.GetProtocolPlugin().GetProtocolByName("tgn52")
	if err != nil {
		return
	}
	t.Log(p.Info().Name)
	t.Log(p.Info().Types)
	t.Log(p.Info().HandleType)
	t.Log(p.Info())
}

type TestData struct {
	Name  string
	Value string
}

// 测试协议的编码方法
func TestProtocolEncode(t *testing.T) {
	p, err := plugins.GetProtocolPlugin().GetProtocolByName("tf100")
	if err != nil {
		t.Fatal(err.Error())
	}
	td := new(TestData)
	td.Name = "aaaa"
	td.Value = "bbbbb"

	var reqData = model.DataReq{}
	reqData.Data = gconv.Bytes(td)
	res := p.Encode(reqData)
	t.Log(res)
}

// 测试自定义协议解析
func TestProtocol(t *testing.T) {
	data := gconv.Bytes("NB1;1234567;1;2;+25.5;00;030;+21;+22")
	p, err := plugins.GetProtocolPlugin().GetProtocolByName("tgn52")
	if err != nil {
		return
	}
	var dr model.DataReq
	dr.Data = data
	res := p.Decode(dr)

	t.Log(res)
}

// 测试插件服务使用，需要先将要测试的插件进行编译
func TestProtocolPluginServer(t *testing.T) {
	plugins.GetProtocolPlugin()
	NetData()
}

func NetData() {
	fmt.Println("Starting the server ...")
	// 创建 listener
	listener, err := net.Listen("tcp", "localhost:5000")
	if err != nil {
		fmt.Println("Error listening", err.Error())
		return //终止程序
	}
	// 监听并接受来自客户端的连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting", err.Error())
			return // 终止程序
		}
		go doServerStuff(conn)
	}
}

func doServerStuff(conn net.Conn) {
	//获取插件
	//pm := GetPlugin(ProtocolPluginName)

	for {
		buf := make([]byte, 512)
		l, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading", err.Error())
			return //终止程序
		}
		fmt.Printf("Received data: %v\n", string(buf[:l]))

		//获取协议插件解析后的数据 传入插件ID，及需要解析的数据
		data, err := plugins.GetProtocolPlugin().GetProtocolDecodeData("tgn52", buf[:l])
		fmt.Println("============通过插件获取数据：", data)
	}
}

func TestNotice(t *testing.T) {

	// 准备通知数据
	var nso []model.NoticeSendObject
	nso = append(nso, model.NoticeSendObject{
		Name:  "mail",
		Value: "xjy@sagoo.cn",
	})
	nso = append(nso, model.NoticeSendObject{
		Name:  "wework",
		Value: "all",
	})

	var msg = model.NoticeInfoData{}
	msg.Totag = nso
	msg.MsgBody = "{'code':'19001'}"
	msg.MsgTitle = "title111112222"
	msg.TemplateCode = "SMS_464050874"

	//通过邮件发送通知
	res, err := plugins.GetNoticePlugin().NoticeSend("mail", msg)
	if err != nil {
		t.Log("Error: ", err.Error())
	}
	t.Log(res)
	//
	////通过短信发送通知
	//res, err := plugins.GetNoticePlugin().NoticeSend("sms", msg)
	//if err != nil {
	//	t.Log("Error: ", err.Error())
	//}
	//t.Log(res)
	//
	//通过webhook发送通知
	//res, err := plugins.GetNoticePlugin().NoticeSend("webhook", msg)
	//if err != nil {
	//	t.Log("Error: ", err.Error())
	//}
	//t.Log(res)
	//
	////通过企业微信发送通知
	//res, err := plugins.GetNoticePlugin().NoticeSend("wework", msg)
	//if err != nil {
	//	t.Log("Error: ", err.Error())
	//}
	//t.Log(res)
	//
	////通过钉钉发送通知
	//res, err = plugins.GetNoticePlugin().NoticeSend("dingding", msg)
	//if err != nil {
	//	t.Log("Error: ", err.Error())
	//}
	//t.Log(res)

}

// TestModbus 测试modbus
func TestModbus(t *testing.T) {
	data := gconv.Bytes("aadsfsfsfdfsfsdfsfs")
	res, err := plugins.GetProtocolPlugin().GetProtocolDecodeData("modbus", data)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log(res)
}
