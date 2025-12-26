package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	sysv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/util/repo"
	utilstring "olares.com/backup-server/pkg/util/string"
)

func CheckSnapshotNotifyState(snapshot *sysv1.Snapshot, field string) (bool, error) {
	if snapshot.Spec.Extra == nil {
		return false, fmt.Errorf("snapshot %s extra is nil", snapshot.Name)
	}

	notifyState, ok := snapshot.Spec.Extra["push"]
	if !ok {
		return false, fmt.Errorf("snapshot %s extra push is nil", snapshot.Name)
	}

	var s *SnapshotNotifyState
	if err := json.Unmarshal([]byte(notifyState), &s); err != nil {
		return false, err
	}

	switch field {
	case "progress":
		return s.Progress, nil
	case "result":
		return s.Result, nil
	case "prepare":
		return s.Prepare, nil
	}

	return false, fmt.Errorf("field not found")
}

func GetBackupPassword(ctx context.Context, owner string, backupName string, token string) (string, error) {
	var pwdResp *passwordResponse
	var backoff = wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    3,
	}

	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		settingsUrl := fmt.Sprintf("http://settings.user-system-%s:28080/api/backup/password", owner)
		client := resty.New().SetTimeout(15 * time.Second).SetDebug(true)

		req := &proxyRequest{
			Op:       "getAccount",
			DataType: "backupPassword",
			Version:  "v1",
			Group:    "service.settings",
			Data:     backupName,
		}

		log.Info("fetch password from settings, ", settingsUrl)
		resp, err := client.R().SetContext(ctx).
			SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
			SetHeader("Authorization", fmt.Sprintf("Bearer %s", token)).
			SetBody(req).
			SetResult(&passwordResponse{}).
			Post(settingsUrl)

		if err != nil {
			log.Error("request settings password api error, ", err)
			return errors.New("get backup password error")
		}

		if resp.StatusCode() != http.StatusOK {
			log.Error("request settings password api response not ok, ", resp.StatusCode())
			return fmt.Errorf("get backup password failed, http status code: %d", resp.StatusCode())
		}

		pwdResp = resp.Result().(*passwordResponse)
		log.Infof("settings password api response, %+v", pwdResp)
		if pwdResp.Code != 0 {
			log.Error("request settings password api response error, ", pwdResp.Code, ", ", pwdResp.Message)
			if strings.Contains(pwdResp.Header.Message, "Secret not found") {
				return fmt.Errorf("get backup password failed, code: %d, message: password not found", pwdResp.Header.Code)
			}
			return fmt.Errorf("get backup password failed, code: %d, message: %s", pwdResp.Header.Code, pwdResp.Header.Message)
		}

		if pwdResp.Data == nil {
			log.Error("request settings password api response data is nil, ", pwdResp.Code, ", ", pwdResp.Message)
			return fmt.Errorf("get backup password, not exists")
		}
		return nil
	}); err != nil {
		return "", err
	}

	return pwdResp.Data.Value, nil
}

func GetClusterId() (string, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var clusterId string
	factory, err := client.NewFactory()
	if err != nil {
		return clusterId, errors.WithStack(err)
	}

	dynamicClient, err := factory.DynamicClient()
	if err != nil {
		return clusterId, errors.WithStack(err)
	}

	var backoff = wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    5,
	}

	if err := retry.OnError(backoff, func(err error) bool {
		return true
	}, func() error {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		unstructuredUser, err := dynamicClient.Resource(constant.OlaresGVR).Get(ctx, constant.OlaresName, metav1.GetOptions{})
		if err != nil {
			return errors.WithStack(err)
		}
		obj := unstructuredUser.UnstructuredContent()
		clusterId, _, err = unstructured.NestedString(obj, "metadata", "labels", "bytetrade.io/cluster-id")
		if err != nil {
			return errors.WithStack(err)
		}
		if clusterId == "" {
			return errors.WithStack(fmt.Errorf("cluster id not found"))
		}
		return nil
	}); err != nil {
		return clusterId, errors.WithStack(err)
	}

	return clusterId, nil
}

func ParseSnapshotTypeText(snapshotType *int) string {
	var t = *snapshotType
	switch t {
	case 0:
		return constant.FullyBackup
	case 1:
		return constant.IncrementalBackup
	default:
		return constant.UnKnownBackup
	}
}

func ParseSnapshotType(snapshotType string) *int {
	var r = constant.UnKnownBackupId
	switch snapshotType {
	case constant.FullyBackup:
		r = constant.FullyBackupId
	case constant.IncrementalBackup:
		r = constant.IncrementalBackupId
	}
	return &r
}

func ParseSnapshotTypeTitle(snapshotType *int) string {
	var t = constant.UnKnownBackup

	if snapshotType == nil || (*snapshotType < 0 || *snapshotType > 1) {
		return utilstring.Title(t)
	}
	if *snapshotType == 0 {
		t = constant.FullyBackup
	} else {
		t = constant.IncrementalBackup
	}

	return utilstring.Title(t)
}

func GetBackupPath(backup *sysv1.Backup) string {
	var p string
	for k, v := range backup.Spec.BackupType {
		if k != "file" {
			continue
		}
		var m = make(map[string]string)
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			log.Errorf("unmarshal backup type error: %v, value: %s", err, v)
			continue
		}
		p = m["path"]
	}

	return p
}

func GetRestorePath(restore *sysv1.Restore) string {
	var p string
	for k, v := range restore.Spec.RestoreType {
		if k != "file" {
			continue
		}
		var m = make(map[string]string)
		if err := json.Unmarshal([]byte(v), &m); err != nil {
			log.Errorf("unmarshal backup type error: %v, value: %s", err, v)
			continue
		}
		p = m["path"]
	}

	return p
}

func GetBackupLocationConfig(backup *sysv1.Backup) (map[string]string, error) {
	var locationConfig map[string]string
	var err error

	for k, v := range backup.Spec.Location {
		if err = json.Unmarshal([]byte(v), &locationConfig); err != nil {
			return nil, err
		}
		_, ok := locationConfig["name"]
		if util.ListContains([]string{
			constant.BackupLocationSpace.String(),
			constant.BackupLocationAwsS3.String(),
			constant.BackupLocationTencentCloud.String(),
		}, k) && !ok {
			return nil, fmt.Errorf("location %s config name not exsits, config: %s", k, v)
		}

		_, ok = locationConfig["path"]
		if k == constant.BackupLocationFileSystem.String() && !ok {
			return nil, fmt.Errorf("location %s config path not exsits, config: %s", k, v)
		}

		locationConfig["location"] = k
		break
	}

	return locationConfig, nil
}

func ParseBackupSnapshotFrequency(str string) string {
	str = strings.ReplaceAll(str, "@", "")
	return utilstring.Title(str)
}

func ParseLocationConfig(locationConfig map[string]string) string {
	var location string
	if locationConfig == nil {
		return ParseBackupLocation(location)
	}

	for l, _ := range locationConfig {
		location = l
	}
	return ParseBackupLocation(location)
}

func ParseBackupLocation(l string) string {
	switch l {
	case constant.BackupLocationSpace.String():
		return constant.BackupLocationSpaceAlias.String()
	case constant.BackupLocationAwsS3.String():
		return constant.BackupLocationAwsS3Alias.String()
	case constant.BackupLocationTencentCloud.String():
		return constant.BackupLocationCosAlias.String()
	case constant.BackupLocationFileSystem.String():
		return constant.BackupLocationFileSystemAlias.String()
	default:
		return constant.BackupLocationUnKnownAlias.String()
	}
}

func ParseSnapshotSize(size *uint64) string {
	if size == nil {
		return "0"
	}

	return fmt.Sprintf("%d", *size)
}

func ParseBackupTypePath(backupType map[string]string) string {
	if backupType == nil {
		return ""
	}
	var backupTypeValue map[string]string
	for k, v := range backupType {
		if k != "file" {
			continue
		}
		if err := json.Unmarshal([]byte(v), &backupTypeValue); err != nil {
			return ""
		}
		return backupTypeValue["path"]
	}
	return ""
}

func GetNextBackupTime(bp sysv1.BackupPolicy) *int64 {
	var res int64

	timeParts := strings.Split(bp.TimesOfDay, ":")
	if len(timeParts) != 2 {
		return nil
	}

	hours, errHour := strconv.Atoi(timeParts[0])
	minutes, errMin := strconv.Atoi(timeParts[1])
	if errHour != nil || errMin != nil {
		return nil
	}

	switch bp.SnapshotFrequency {
	case "@hourly":
		res = getNextBackupTimeByHourly(minutes).Unix()
	case "@weekly":
		res = getNextBackupTimeByWeekly(hours, minutes, bp.DayOfWeek).Unix()
	case "@monthly":
		res = getNextBackupTimeByMonthly(hours, minutes, bp.DateOfMonth).Unix()
	default:
		res = getNextBackupTimeByDaily(hours, minutes).Unix()
	}
	return &res
}

func getNextBackupTimeByDaily(hours, minutes int) time.Time {
	var utcLocation, _ = time.LoadLocation("")
	var n = time.Now().In(utcLocation)

	today := time.Date(n.Year(), n.Month(), n.Day(), hours, minutes, 0, 0, n.Location())

	if today.Before(n) {
		return today.AddDate(0, 0, 1)
	}

	return today
}

func getNextBackupTimeByMonthly(hours, minutes int, day int) time.Time {
	var fmtDate = "%d-%.2d-%.2d %.2d:%.2d:00"
	var utcLocation, _ = time.LoadLocation("")
	var n = time.Now().In(utcLocation)

	var year = n.Year()
	var month = int(n.Month())

	fDate, err := time.ParseInLocation(util.DateFormat, fmt.Sprintf(fmtDate, year, month, day, hours, minutes), utcLocation)

	if fDate.Before(n) {
		for {
			month++
			if month > 12 {
				month = 1
				year++
			}
			fDate, err = time.ParseInLocation(util.DateFormat, fmt.Sprintf(fmtDate, year, month, day, hours, minutes), utcLocation)
			if err != nil {
				continue
			}
			break
		}
	}

	if hours >= 16 {
		fDate = fDate.AddDate(0, 0, -1)
	}

	return fDate
}

func getNextBackupTimeByWeekly(hours, minutes int, weekly int) time.Time {
	var utcLocation, _ = time.LoadLocation("")
	weekly = weekly - 1
	var n = time.Now().In(utcLocation)
	firstDayOfWeek := util.GetFirstDayOfWeek(n)

	backupDay := firstDayOfWeek.AddDate(0, 0, weekly)

	backupTime := time.Date(backupDay.Year(), backupDay.Month(), backupDay.Day(),
		hours, minutes, 0, 0, n.Location())

	if hours >= 16 {
		backupTime = backupTime.AddDate(0, 0, -1)
	}

	if backupTime.Before(n) {
		backupTime = backupTime.AddDate(0, 0, 7)
	}

	return backupTime
}

func getNextBackupTimeByHourly(minutes int) time.Time {
	now := time.Now()

	currentMinute := now.Minute()
	nextMinute := currentMinute

	remainder := currentMinute % minutes
	if remainder == 0 && now.Second() == 0 && now.Nanosecond() == 0 {
		nextMinute = currentMinute + minutes
	} else {
		nextMinute = currentMinute + (minutes - remainder)
	}

	minutesToAdd := nextMinute - currentMinute

	nextTime := now.Add(time.Duration(minutesToAdd) * time.Minute).
		Truncate(time.Minute)

	return nextTime
}

func GetRestoreType(restore *sysv1.Restore) (string, error) {
	var data = restore.Spec.RestoreType
	_, ok := data[constant.BackupTypeFile]
	if ok {
		return constant.BackupTypeFile, nil
	}

	_, ok = data[constant.BackupTypeApp]
	if ok {
		return constant.BackupTypeApp, nil
	}

	return "", fmt.Errorf("restore from backup type invalid")
}

func ParseRestoreType(backupType string, restore *sysv1.Restore) (*RestoreType, error) {
	var m *RestoreType
	var data = restore.Spec.RestoreType
	v, ok := data[backupType]
	if !ok {
		return nil, errors.WithStack(fmt.Errorf("restore type data not found"))
	}

	if err := json.Unmarshal([]byte(v), &m); err != nil {
		return nil, errors.WithStack(err)
	}
	return m, nil
}

func ParseBackupNameFromRestore(restore *sysv1.Restore) string {
	if restore == nil || restore.Spec.RestoreType == nil {
		return ""
	}

	data, ok := restore.Spec.RestoreType[constant.BackupTypeFile]
	if !ok {
		return ""
	}

	var r *RestoreType
	if err := json.Unmarshal([]byte(data), &r); err != nil {
		return ""
	}
	return r.BackupName
}

func ParseRestoreBackupUrlDetail(owner, u string) (storage *RestoreBackupUrlDetail, backupName, backupId string, resticSnapshotId string, snapshotTime string, location string, err error) {
	if u == "" {
		err = fmt.Errorf("backupUrl is empty")
		return
	}

	u = strings.TrimPrefix(u, "s3:")
	u = strings.TrimRight(u, "/")
	backupUrlType, e := ParseBackupUrl(owner, u)
	if e != nil {
		err = errors.WithMessage(e, fmt.Sprintf("parse backupUrl failed, backupUrl: %s", u))
		return
	}

	log.Infof("backup url type: %s", util.ToJSON(backupUrlType))

	if backupName = backupUrlType.BackupName; backupName == "" {
		err = errors.WithStack(fmt.Errorf("backupName is empty, backupUrl: %s", u))
		return
	}

	if backupId = backupUrlType.BackupId; backupId == "" {
		err = errors.WithStack(fmt.Errorf("backupId is empty, backupUrl: %s", u))
		return
	}

	if resticSnapshotId = backupUrlType.Values.Get("snapshotId"); resticSnapshotId == "" {
		err = errors.WithStack(fmt.Errorf("snapshotId is empty, backupUrl: %s", u))
		return
	}

	if snapshotTime = backupUrlType.Values.Get("snapshotTime"); snapshotTime == "" {
		err = errors.WithStack(fmt.Errorf("snapshotTime is empty, backupUrl: %s", u))
		return
	}

	location = backupUrlType.Location
	storage, err = backupUrlType.GetStorage()
	if err != nil {
		return
	}

	return
}

func IsBackupLocationSpace(u string) bool {
	return strings.Contains(u, constant.LocationTypeSpaceTag)
}

func ParseBackupUrl(owner, s string) (*BackupUrlType, error) {
	userspacePath, _, appcachePath, _, err := GetUserspacePvc(owner)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	u, err := url.Parse(s) // if filesystem, u.Path is like '/Files/Home/...'
	if err != nil {
		return nil, errors.WithStack(err)
	}

	log.Infof("backup url type: %s", u.Path)

	var location string
	if u.Scheme == "" && u.Path[:1] == "/" {
		location = constant.BackupLocationFileSystem.String()
	} else if strings.Contains(u.Path, constant.LocationTypeSpaceTag) {
		location = constant.BackupLocationSpace.String()
	} else if strings.Contains(u.Host, constant.LocationTypeAwsS3Tag) {
		location = constant.BackupLocationAwsS3.String()
	} else if strings.Contains(u.Host, constant.LocationTypeTencentCloudTag) {
		location = constant.BackupLocationTencentCloud.String()
	}

	if location == "" {
		return nil, fmt.Errorf("location is empty, url: %s", s)
	}

	if strings.TrimPrefix(u.Path, "/") == "" {
		return nil, fmt.Errorf("url invalid, path: %s", u.Path)
	}

	idx := strings.Index(u.Path, constant.DefaultStoragePrefix)
	if idx == -1 {
		return nil, fmt.Errorf("url invalid, url: %s", u)
	}

	pathSuffix := u.Path[idx+len(constant.DefaultStoragePrefix):]

	backupName, backupId, err := utilstring.SplitPath(pathSuffix)
	if err != nil {
		return nil, fmt.Errorf("split path errror: %v, url: %s", err, s)
	}

	repoInfo, err := FormatEndpoint(location, u.Scheme, u.Host, u.Path, userspacePath, appcachePath)
	if err != nil {
		return nil, err
	}

	log.Infof("backup url repoInfo.endpoint: %s", repoInfo.Endpoint)

	if location == constant.BackupLocationFileSystem.String() {
		if !util.IsExist(repoInfo.Endpoint) {
			return nil, fmt.Errorf("backup dir not exists, path: %s", repoInfo.Endpoint)
		}
	}

	var cloudName string
	if location == constant.BackupLocationSpace.String() {
		if strings.Contains(u.Host, constant.LocationTypeAwsS3Tag) {
			cloudName = "aws"
		} else if strings.Contains(u.Host, constant.LocationTypeTencentCloudTag) {
			cloudName = constant.BackupLocationTencentCloud.String()
		}
	}

	var res = &BackupUrlType{
		Schema:     u.Scheme,
		Host:       u.Host,
		Path:       strings.TrimRight(u.Path, "/"),
		Values:     u.Query(),
		Location:   location,
		Endpoint:   repoInfo.Endpoint,
		BackupId:   backupId,
		BackupName: backupName,
		PvcPath:    userspacePath,

		CloudName:    cloudName,
		Region:       repoInfo.Region,
		Bucket:       repoInfo.Bucket,
		Prefix:       repoInfo.Prefix,
		OlaresSuffix: repoInfo.Suffix,
	}

	if location == constant.BackupLocationFileSystem.String() {
		res.FilesystemPath = repoInfo.Endpoint
	}
	return res, nil
}

func GetBackupType(backup *sysv1.Backup) string {
	var backupType string
	for k := range backup.Spec.BackupType {
		backupType = k
		break
	}

	return backupType
}

func GetRestoreAppName(restore *sysv1.Restore) string {
	app, ok := restore.Spec.RestoreType["app"]
	if !ok {
		return ""
	}
	var data = make(map[string]interface{})
	if err := json.Unmarshal([]byte(app), &data); err != nil {
		return ""
	}

	appName, ok := data["name"]
	if !ok {
		return ""
	}

	return appName.(string)
}

func GetBackupAppName(backup *sysv1.Backup) string {
	app, ok := backup.Spec.BackupType["app"]
	if !ok {
		return ""
	}

	var data = make(map[string]string)
	if err := json.Unmarshal([]byte(app), &data); err != nil {
		return ""
	}

	appName, ok := data["name"]
	if !ok {
		return ""
	}

	return appName
}

func GenericPager[T runtime.Object](limit int64, offset int64, resourceList T) (T, int64, int64) {
	if limit <= 0 {
		limit = 5
	}
	if offset < 0 {
		offset = 0
	}

	listValue := reflect.ValueOf(resourceList)
	if listValue.Kind() == reflect.Ptr {
		listValue = listValue.Elem()
	}

	itemsField := listValue.FieldByName("Items")
	if !itemsField.IsValid() || itemsField.Kind() != reflect.Slice {
		return resourceList, 1, 1
	}

	total := int64(itemsField.Len())
	resultList := reflect.New(reflect.TypeOf(resourceList).Elem()).Elem()

	if typeMetaField := resultList.FieldByName("TypeMeta"); typeMetaField.IsValid() {
		originalTypeMetaField := listValue.FieldByName("TypeMeta")
		if originalTypeMetaField.IsValid() {
			typeMetaField.Set(originalTypeMetaField)
		}
	}

	if listMetaField := resultList.FieldByName("ListMeta"); listMetaField.IsValid() {
		originalListMetaField := listValue.FieldByName("ListMeta")
		if originalListMetaField.IsValid() {
			listMetaField.Set(originalListMetaField)
		}
	}

	startIndex := offset
	endIndex := offset + limit

	if startIndex >= total {
		emptySlice := reflect.MakeSlice(itemsField.Type(), 0, 0)
		resultList.FieldByName("Items").Set(emptySlice)
	} else {
		if endIndex > total {
			endIndex = total
		}

		newItemsSlice := reflect.MakeSlice(itemsField.Type(), int(endIndex-startIndex), int(endIndex-startIndex))
		for i := startIndex; i < endIndex; i++ {
			newItemsSlice.Index(int(i - startIndex)).Set(itemsField.Index(int(i)))
		}
		resultList.FieldByName("Items").Set(newItemsSlice)
	}

	currentPage := offset/limit + 1
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return resultList.Addr().Interface().(T), currentPage, totalPages
}

func GetUserspacePvc(owner string) (string, string, string, string, error) {
	f, err := client.NewFactory()
	if err != nil {
		return "", "", "", "", errors.WithStack(err)
	}

	c, err := f.KubeClient()
	if err != nil {
		return "", "", "", "", errors.WithStack(err)
	}

	res, err := c.AppsV1().StatefulSets("user-space-"+owner).Get(context.TODO(), "bfl", metav1.GetOptions{})
	if err != nil {
		return "", "", "", "", errors.Wrap(err, fmt.Sprintf("get bfl failed, owner: %s", owner))
	}

	userspacePvc, ok := res.Annotations["userspace_pvc"]
	if !ok {
		return "", "", "", "", fmt.Errorf("bfl userspace_pvc not found, owner: %s", owner)
	}

	appcachePvc, ok := res.Annotations["appcache_pvc"]
	if !ok {
		return "", "", "", "", fmt.Errorf("bfl appcache_pvc not found, owner: %s", owner)
	}

	var p = path.Join("/", "rootfs", "userspace", userspacePvc)
	var cachePath = path.Join("/", "appcache", "Cache", appcachePvc) // /appcache/Cache/xxx

	return p, userspacePvc, cachePath, appcachePvc, nil
}

func TrimPathPrefix(p string) (bool, bool, string) {
	if !strings.HasSuffix(p, "/") {
		p = p + "/"
	}
	var external = fmt.Sprintf("/Files/External/%s/", constant.NodeName)
	var cache = fmt.Sprintf("/Cache/%s/", constant.NodeName)
	if strings.HasPrefix(p, external) {
		return true, false, strings.TrimPrefix(p, external)
	} else if strings.HasPrefix(p, cache) {
		return false, true, strings.TrimPrefix(p, cache)
	} else if strings.HasPrefix(p, "/Files/") {
		return false, false, strings.TrimPrefix(p, "/Files/")
	} else {
		return false, false, p
	}
}

// used for restore
func FormatEndpoint(location, schema, host, urlPath, pvc, cachepvc string) (*repo.RepositoryInfo, error) {
	var p = utilstring.TrimSuffix(urlPath, constant.DefaultStoragePrefix)

	switch location {
	case constant.BackupLocationSpace.String():
		return repo.FormatSpace(schema, host, p)
	case constant.BackupLocationTencentCloud.String():
		return repo.FormatCos(schema, host, p)
	case constant.BackupLocationAwsS3.String():
		var rawUrl = fmt.Sprintf("%s://%s%s", schema, host, p)
		result, err := repo.FormatS3(rawUrl)
		if err != nil {
			return nil, err
		}
		result.Endpoint = strings.TrimPrefix(result.Endpoint, "s3:")
		return result, nil
	case constant.BackupLocationFileSystem.String():
		external, cache, relativePath := TrimPathPrefix(p)
		var endpoint string
		if external {
			endpoint = path.Join(constant.ExternalPath, relativePath)
		} else if cache {
			endpoint = path.Join(cachepvc, relativePath)
		} else {
			endpoint = path.Join(pvc, relativePath)
		}
		return &repo.RepositoryInfo{
			Endpoint: endpoint,
		}, nil
	default:
		return nil, fmt.Errorf("location invalid: %s", location)
	}
}

func GetBackupTypeFromTags(tags []string) (backupType string) {
	backupType = constant.BackupTypeFile
	if tags == nil || len(tags) == 0 {
		return
	}

	for _, tag := range tags {
		e := strings.Index(tag, "=")
		if e >= 0 {
			if tag[:e] == "backup-type" {
				backupType = tag[e+1:]
				break
			}
		}
	}

	return
}

func GetBackupTypeAppName(tags []string) (backupTypeAppName string) {
	if tags == nil || len(tags) == 0 {
		return
	}
	for _, tag := range tags {
		e := strings.Index(tag, "=")
		if e >= 0 {
			if tag[:e] == "backup-app-type-name" {
				backupTypeAppName = tag[e+1:]
				break
			}
		}
	}

	if backupTypeAppName != "" {
		tmp, err := util.Base64decode(backupTypeAppName)
		if err != nil {
			return ""
		}
		backupTypeAppName = string(tmp)
	}

	return
}

func GetBackupTypeFilePath(tags []string) (backupPath string) {
	if tags == nil || len(tags) == 0 {
		return
	}

	for _, tag := range tags {
		e := strings.Index(tag, "=")
		if e >= 0 {
			if tag[:e] == "backup-path" {
				backupPath = tag[e+1:]
				break
			}
		}
	}

	if backupPath != "" {
		tmp, err := util.Base64decode(backupPath)
		if err != nil {
			return ""
		}
		backupPath = string(tmp)
	}

	return
}
