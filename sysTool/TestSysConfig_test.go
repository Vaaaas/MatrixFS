package sysTool

import(
 	"testing"

	"github.com/golang/glog"
)

func TestInitConfig(t *testing.T) {
	InitConfig(-1,-1)
	glog.Infof("系统参数配置：容错数 %d , 行数 %d , 数据分块数 %d , 数据列数 %d , 冗余列数 %d", SysConfig.FaultNum, SysConfig.RowNum, SysConfig.SliceNum, SysConfig.DataNum, SysConfig.RddtNum)
}