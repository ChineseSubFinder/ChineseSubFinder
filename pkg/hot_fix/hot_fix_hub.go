package hot_fix

import (
	"errors"
	"fmt"

	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/ifaces"
	"github.com/ChineseSubFinder/ChineseSubFinder/pkg/types"

	"github.com/ChineseSubFinder/ChineseSubFinder/internal/dao"
	"github.com/ChineseSubFinder/ChineseSubFinder/internal/models"
	"github.com/sirupsen/logrus"
)

// HotFixProcess 去 DB 中查询 Hotfix 的标记，看有那些需要修复，那些已经修复完毕
func HotFixProcess(log *logrus.Logger, param types.HotFixParam) error {

	// -----------------------------------------------------------------------
	// 一共有多少个 HotFix 要修复，需要固定下来
	hotfixCases := []ifaces.IHotFix{
		NewHotFix001(log, param.MovieRootDirs, param.SeriesRootDirs),
		NewHotFix002(log),
		NewHotFix003(log),
		NewHotFix004(log),
		NewHotFix005(log),
		NewHotFix006(log),
		NewHotFix007(log),
		// 注意下面的 switch case 也要相应的加
	}
	// -----------------------------------------------------------------------
	// 找现在有多少个 hotfix 执行过了
	var hotFixes []models.HotFix
	result := dao.GetDb().Find(&hotFixes)
	if result == nil || result.Error != nil {
		return errors.New(fmt.Sprintf("hotfix query all result failed"))
	}
	// 数据库中是否有记录，记录了是否有运行都需要判断
	var hotFixRecord = make(map[string]models.HotFix)
	for _, fix := range hotFixes {
		hotFixRecord[fix.Key] = fix
	}
	// 交叉对比，这个执行的顺序又上面 []ifaces.IHotFix 指定
	for _, hotfixCase := range hotfixCases {
		_, bFound := hotFixRecord[hotfixCase.GetKey()]
		if bFound == true && hotFixRecord[hotfixCase.GetKey()].Done == true {
			// 如果修复过了，那么就跳过
			continue
		}
		// 没有找到那么就需要进行修复
		processResult, err := hotfixCase.Process()
		// 找到对应的 hotfix 方案进行 interface 数据的转换输出
		switch hotfixCase.GetKey() {
		case "001":
			outStruct := processResult.(OutStruct001)
			if err != nil {
				for i, file := range outStruct.ErrFiles {
					log.Errorln("Hotfix 001, rename failed,", i, file)
				}
				// 如果任意故障则跳出后续的修复
				log.Errorln("Hotfix 001 failed, break")
				return err
			} else {
				for i, file := range outStruct.RenamedFiles {
					log.Infoln("Hotfix 001, rename done,", i, file)
				}
			}
			break
		case "002":
			log.Infoln("Hotfix 002, process == ", processResult.(bool))
			break
		case "003":
			log.Infoln("Hotfix 003, process == ", processResult.(bool))
			break
		case "004":
			log.Infoln("Hotfix 004, process == ", processResult.(bool))
			break
		case "005":
			log.Infoln("Hotfix 005, process == ", processResult.(bool))
			break
		case "006":
			log.Infoln("Hotfix 006, process == ", processResult.(bool))
			break
		case "007":
			log.Infoln("Hotfix 007, process == ", processResult.(bool))
			break
		default:
			continue
		}

		var hotfixs []models.HotFix
		dao.GetDb().Where("key = ?", hotfixCase.GetKey()).Find(&hotfixs)
		if len(hotfixs) < 1 {
			// 不存在则新建
			// 执行成功则存入数据库中，标记完成
			markHotFixDone := models.HotFix{Key: hotfixCase.GetKey(), Done: true}
			result = dao.GetDb().Create(&markHotFixDone)
			if result == nil {
				nowError := errors.New(fmt.Sprintf("hotfix %s is done, but record failed, dao.GetDb().Create return nil", hotfixCase.GetKey()))
				log.Errorln(nowError)
				return nowError
			}
			if result.Error != nil {
				nowError := errors.New(fmt.Sprintf("hotfix %s is done, but record failed, %s", hotfixCase.GetKey(), result.Error))
				log.Errorln(nowError)
				return nowError
			}
			log.Infoln("Hotfix", hotfixCase.GetKey(), "is Recorded")
			// 找到了，目前的逻辑是成功才插入，那么查询到了，就默认是执行成功了
		} else {
			// 存在则更新
			hotfixs[0].Done = true
			dao.GetDb().Save(&hotfixs[0])
		}

	}
	return nil
}
