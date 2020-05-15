package httpd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/miiniper/tgmsg_bot"

	"k8s.io/api/apps/v1beta1"
	v1 "k8s.io/api/core/v1"

	"github.com/miiniper/loges"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"

	"github.com/julienschmidt/httprouter"
)

type K8sConfig struct {
	ClusterName string `json:"clustername"`
	ConfigFile  string `json:"configfile"`
}

type K8sConfigs []K8sConfig

type HttpStatus struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

var Bot tgmsg_bot.Bot
var ClusterCfgs K8sConfigs

func Init() {
	Bot = tgmsg_bot.NewBot("msg")
	ClusterCfgs = GetConfig()
}

func (s *Service) PodsCheck(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	for _, ClusterCfg := range ClusterCfgs {
		cli, _ := K8sCli(ClusterCfg.ConfigFile)
		pods, _ := GetPod(cli)

		for _, j := range pods.Items {
			if j.Status.ContainerStatuses[0].Ready != true {
				msg := fmt.Sprintf("cluster :%s\n ns: %s\n pod: %s\n is notReady", ClusterCfg.ClusterName, j.ObjectMeta.Namespace, j.ObjectMeta.Name)
				Bot.SendMsg(msg)
			}
		}

	}
	w.Write([]byte("ok"))
}

func (s *Service) DepCheck(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	for _, ClusterCfg := range ClusterCfgs {
		cli, _ := K8sCli(ClusterCfg.ConfigFile)
		deps, _ := GetDeployment(cli)

		for _, j := range deps.Items {
			go func() {
				if *j.Spec.Replicas != j.Status.ReadyReplicas {
					msg := fmt.Sprintf("cluster :%s\n ns: %s\n dep: %s\n some pod is  notReady", ClusterCfg.ClusterName, j.ObjectMeta.Namespace, j.ObjectMeta.Name)
					Bot.SendMsg(msg)
				}
			}()

		}

	}
	w.Write([]byte("ok"))
}

func K8sCli(k8sCfg string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(k8sCfg))
	if err != nil {
		loges.Loges.Error("REST Config From KubeConfig is err:", zap.Error(err))
		return nil, err
	}

	cli, err := kubernetes.NewForConfig(config)
	if err != nil {
		loges.Loges.Error("new KubeConfig is err:", zap.Error(err))
		return nil, err
	}
	return cli, nil

}

func GetPod(cli *kubernetes.Clientset) (*v1.PodList, error) {
	podAll, err := cli.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		loges.Loges.Error("get pod info  is err:", zap.Error(err))
		return nil, err
	}
	return podAll, nil

}

func GetDeployment(cli *kubernetes.Clientset) (*v1beta1.DeploymentList, error) {
	dep, err := cli.AppsV1beta1().Deployments("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		loges.Loges.Error("get pod info  is err:", zap.Error(err))
		return nil, err
	}
	return dep, nil
}

func GetConfig() K8sConfigs {
	session, err := mgo.Dial(viper.GetString("db.addr"))
	if err != nil {
		loges.Loges.Error("conn mgo is err:", zap.Error(err))
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	err = session.DB("admin").Login(viper.GetString("db.dbuser"), viper.GetString("db.dbpass"))
	if err != nil {
		loges.Loges.Error("auth mgo is err:", zap.Error(err))
	}
	aa := K8sConfigs{}
	c := session.DB("check").C("k8sconfig")
	err = c.Find(nil).All(&aa)
	if err != nil {
		loges.Loges.Error("select db is err:", zap.Error(err))
	}

	return aa
}

//
//func AddConfig() {
//	bb, err := ioutil.ReadFile("/home/han/config")
//	if err != nil {
//		fmt.Println("111111111", err)
//	}
//	k8sc := K8sConfig{}
//	k8sc.ClusterName = "tencent-c"
//	k8sc.ConfigFile = string(bb)
//
//	//fmt.Println(k8sc)
//
//	session, err := mgo.Dial(viper.GetString("db.addr"))
//	if err != nil {
//		fmt.Println("333333333333333333333", err)
//	}
//	defer session.Close()
//	session.SetMode(mgo.Monotonic, true)
//	err = session.DB("admin").Login(viper.GetString("db.dbuser"), viper.GetString("db.dbpass"))
//	if err != nil {
//		fmt.Println("2222222222222222222", err)
//	}
//	c := session.DB("check").C("k8sconfig")
//	err = c.Insert(&k8sc)
//	if err != nil {
//		fmt.Println("44444444444444444444", err)
//	}
//}
