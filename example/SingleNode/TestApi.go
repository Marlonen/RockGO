package main

import (
	"fmt"
	"github.com/zllangct/RockGO/cluster"
	"github.com/zllangct/RockGO/component"
	"github.com/zllangct/RockGO/network"
	"github.com/zllangct/RockGO/network/messageProtocol"
)


/*
协议接口组,需继承network.ApiBase 基类

api结构体中，符合条件的方法会被作为api对客户端服务
api方法的规则为：func 函数名（sess **network.Session, message *消息结构体）
*/

type TestApi struct {
	network.ApiBase
	nodeComponent *Cluster.NodeComponent
}

//使用协议接口时，需先初始化，初始化时需传入定义的消息号对应字典
//以及所需的消息序列化组件，可轻易切换为protobuf，msgpack等其他序列化工具
func NewTestApi(parent *Component.Object) *TestApi  {
	r:=&TestApi{}
	r.Init(r,parent,Testid2mt,&MessageProtocol.JsonProtocol{})
	return r
}

//获取node组件
func (this *TestApi) getNodeComponent()(*Cluster.NodeComponent,error){
	if this.nodeComponent == nil{
		o,err:= this.GetParent()
		if err!=nil {
			return nil,err
		}
		err= o.Root().Find(&this.nodeComponent)
		if err!=nil {
			return nil,err
		}
		return this.nodeComponent,nil
	}
	return this.nodeComponent,nil
}

//协议接口 1  Hello
func (this *TestApi)Hello(sess *network.Session,message *TestMessage) error {
	println(fmt.Sprintf("Hello,%s",message.Name))

	//reply
	err:=sess.Emit(1,[]byte(fmt.Sprintf("hello client %s",message.Name)))
	if err!=nil {
		return err
	}
	return nil
}

//协议接口 2 登录
func (this *TestApi)Login(sess *network.Session,message *TestLogin) error {
	println(fmt.Sprintf("received a client login request,%s ",message.Account))

	//获取node组件
	nodec,err:=this.getNodeComponent()
	if err!=nil {
		return err
	}
	//选择一个该APP的角色节点，可选参数可设置selector，在多个同一角色的节点中选择符合条件的节点
	//可使用随机选择器、最小负载选择器，同时可自定义其他选择器，比如 按地区选择
	node,err:=nodec.GetNode("login",Cluster.SELECTOR_TYPE_MIN_LOAD)
	if err != nil {
		return err
	}
	//通过节点获取rpc客户端
	loginNode,err:= node.GetClient()
	if err!=nil {
		return err
	}
	//调用登录服的登录接口
	var pInfo *PlayerInfo
	err= loginNode.Call("LoginComponent.Login",message.Account,&pInfo)
	if err!=nil {
		return err
	}
	//reply 登录结果反馈到客户端
	err=this.Reply(sess,pInfo)
	if err!=nil {
		return err
	}
	return err
}