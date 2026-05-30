package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"yatori-go-console/entity/dto"
	"yatori-go-console/global"
	"yatori-go-console/web/activity"

	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

const autoExecutionScheduleInterval = time.Minute

var executionTimePattern = regexp.MustCompile(`^(?:[01]\d|2[0-3]):[0-5]\d$`)

func normalizeExecutionTime(value string) string {
	return strings.TrimSpace(value)
}

func validateAutoExecutionWindow(start string, end string) error {
	start = normalizeExecutionTime(start)
	end = normalizeExecutionTime(end)
	if start == "" && end == "" {
		return nil
	}
	if start == "" || end == "" {
		return fmt.Errorf("自动执行时间段需要同时设置开始时间和结束时间")
	}
	if !executionTimePattern.MatchString(start) || !executionTimePattern.MatchString(end) {
		return fmt.Errorf("自动执行时间格式必须为 HH:MM")
	}
	return nil
}

func parseExecutionClockMinutes(value string) (int, error) {
	if !executionTimePattern.MatchString(value) {
		return 0, fmt.Errorf("invalid time: %s", value)
	}
	hour := int(value[0]-'0')*10 + int(value[1]-'0')
	minute := int(value[3]-'0')*10 + int(value[4]-'0')
	return hour*60 + minute, nil
}

func isWithinAutoExecutionWindow(start string, end string, now time.Time) (bool, error) {
	if err := validateAutoExecutionWindow(start, end); err != nil {
		return false, err
	}
	start = normalizeExecutionTime(start)
	end = normalizeExecutionTime(end)
	if start == "" && end == "" {
		return false, nil
	}

	startMinutes, err := parseExecutionClockMinutes(start)
	if err != nil {
		return false, err
	}
	endMinutes, err := parseExecutionClockMinutes(end)
	if err != nil {
		return false, err
	}
	nowMinutes := now.Hour()*60 + now.Minute()

	if startMinutes == endMinutes {
		return true, nil
	}
	if startMinutes < endMinutes {
		return nowMinutes >= startMinutes && nowMinutes < endMinutes, nil
	}
	return nowMinutes >= startMinutes || nowMinutes < endMinutes, nil
}

func shouldManageAutoExecution(user dto.ConfigManagerUser) bool {
	return strings.EqualFold(strings.TrimSpace(user.AccountType), "XUEXITONG")
}

func isActivityRunningInstance(userActivity *activity.Activity) bool {
	if userActivity == nil {
		return false
	}
	if xxt, ok := (*userActivity).(*activity.XXTActivity); ok {
		return xxt.IsRunning
	}
	if yinghua, ok := (*userActivity).(*activity.YingHuaActivity); ok {
		return yinghua.IsRunning
	}
	return false
}

func startManagedActivity(user dto.ConfigManagerUser, userActivity *activity.Activity) {
	go func() {
		if err := (*userActivity).Start(); err != nil {
			lg.Print(lg.INFO, "[", lg.Yellow, user.Account, lg.Default, "] 自动执行启动失败: ", err.Error())
		}
	}()
}

func syncAutoExecutionSchedules(now time.Time) error {
	users, err := fetchConfigManagerUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		start := user.CoursesCustom.AutoRunStartTime
		end := user.CoursesCustom.AutoRunEndTime
		if strings.TrimSpace(start) == "" && strings.TrimSpace(end) == "" {
			continue
		}
		if !shouldManageAutoExecution(user) {
			continue
		}

		withinWindow, err := isWithinAutoExecutionWindow(start, end, now)
		if err != nil {
			lg.Print(lg.INFO, "[", lg.Yellow, user.Account, lg.Default, "] 自动执行时间配置无效: ", err.Error())
			continue
		}

		userPO, err := configManagerUserToPO(user)
		if err != nil {
			lg.Print(lg.INFO, "[", lg.Yellow, user.Account, lg.Default, "] 自动执行账号配置转换失败: ", err.Error())
			continue
		}

		userActivity := global.GetUserActivity(*userPO)
		if userActivity == nil && withinWindow {
			createdActivity := activity.BuildUserActivity(*userPO)
			if createdActivity == nil {
				continue
			}
			userActivity = &createdActivity
			global.PutUserActivity(*userPO, &createdActivity)
		}
		if userActivity == nil {
			continue
		}

		isRunning := isActivityRunningInstance(userActivity)
		if withinWindow && !isRunning {
			startManagedActivity(user, userActivity)
			continue
		}
		if !withinWindow && isRunning {
			if err := (*userActivity).Stop(); err != nil {
				lg.Print(lg.INFO, "[", lg.Yellow, user.Account, lg.Default, "] 自动执行停止失败: ", err.Error())
			}
		}
	}

	return nil
}

func StartAutoExecutionScheduler() {
	go func() {
		if err := syncAutoExecutionSchedules(time.Now()); err != nil {
			lg.Print(lg.INFO, lg.Yellow, "自动执行初始化检查失败: ", err.Error())
		}

		ticker := time.NewTicker(autoExecutionScheduleInterval)
		defer ticker.Stop()
		for now := range ticker.C {
			if err := syncAutoExecutionSchedules(now); err != nil {
				lg.Print(lg.INFO, lg.Yellow, "自动执行巡检失败: ", err.Error())
			}
		}
	}()
}
