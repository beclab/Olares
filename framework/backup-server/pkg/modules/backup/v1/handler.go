package v1

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	sysv1 "olares.com/backup-server/pkg/apis/sys.bytetrade.io/v1"
	"olares.com/backup-server/pkg/apiserver/config"
	"olares.com/backup-server/pkg/apiserver/response"
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/handlers"
	"olares.com/backup-server/pkg/storage"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	backupssdkrestic "olares.com/backups-sdk/pkg/restic"
)

type Handler struct {
	cfg     *config.Config
	factory client.Factory
	handler handlers.Interface
}

func New(cfg *config.Config, factory client.Factory, handler handlers.Interface) *Handler {
	return &Handler{
		cfg:     cfg,
		factory: factory,
		handler: handler,
	}
}

func (h *Handler) health(_ *restful.Request, resp *restful.Response) {
	response.SuccessNoData(resp)
}

func (h *Handler) ready(req *restful.Request, resp *restful.Response) {
	resp.Write([]byte("ok"))
}

func (h *Handler) listBackup(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	owner := req.HeaderParameter(constant.BflUserKey)
	limit := util.ParseToInt64(req.QueryParameter("limit"))
	offset := util.ParseToInt64(req.QueryParameter("offset"))

	backups, err := h.handler.GetBackupHandler().ListBackups(ctx, owner, offset, limit, fmt.Sprintf("owner=%s", owner), "")
	if err != nil {
		log.Errorf("get backups error: %v", err)
		response.Success(resp, parseResponseBackupList(backups, nil, 1))
		return
	}

	result, _, totalPage := handlers.GenericPager(limit, offset, backups)

	labelsSelector := h.handler.GetBackupHandler().GetBackupIdForLabels(result)
	var allSnapshots = new(sysv1.SnapshotList)
	for _, ls := range labelsSelector {
		snapshots, err := h.handler.GetSnapshotHandler().ListSnapshots(ctx, 0, 0, ls, "")
		if err != nil {
			log.Errorf("get snapshots error: %v", err)
			continue
		}
		if snapshots == nil || len(snapshots.Items) == 0 {
			continue
		}
		allSnapshots.Items = append(allSnapshots.Items, snapshots.Items...)
	}

	response.Success(resp, parseResponseBackupList(result, allSnapshots, totalPage))
}

func (h *Handler) get(req *restful.Request, resp *restful.Response) {
	ctx, backupId := req.Request.Context(), req.PathParameter("id")
	owner := req.HeaderParameter(constant.BflUserKey)
	_ = owner

	backup, err := h.handler.GetBackupHandler().GetById(ctx, backupId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessage(err, fmt.Sprintf("get backup %s error", backupId)))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("backup %s not found", backupId))
		return
	}

	response.Success(resp, parseResponseBackupDetail(backup))
}

func (h *Handler) getBackupPlan(req *restful.Request, resp *restful.Response) {
	ctx, backupId := req.Request.Context(), req.PathParameter("id")
	owner := req.HeaderParameter(constant.BflUserKey)
	_ = owner

	backup, err := h.handler.GetBackupHandler().GetById(ctx, backupId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessage(err, fmt.Sprintf("get backup %s error", backupId)))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("backup %s not found", backupId))
		return
	}

	var snapshotSelectLabel = fmt.Sprintf("backup-id=%s", backup.Name)
	var snapshot *sysv1.Snapshot

	snapshots, _ := h.handler.GetSnapshotHandler().ListSnapshots(ctx, 0, 0, snapshotSelectLabel, "")

	if snapshots != nil && len(snapshots.Items) > 0 {
		snapshot = &snapshots.Items[0]
	}

	result, err := parseResponseBackupOne(backup, snapshot)
	if err != nil {
		response.HandleError(resp, fmt.Errorf("backup %s %v", backup.Name, err))
		return
	}

	response.Success(resp, result)
}

func (h *Handler) addBackup(req *restful.Request, resp *restful.Response) {
	var (
		err error
		b   BackupCreate
	)

	if err = req.ReadEntity(&b); err != nil {
		response.HandleError(resp, errors.WithStack(err))
		return
	}

	ctx := req.Request.Context()
	owner := req.HeaderParameter(constant.BflUserKey)

	log.Infof("received backup create request: %s", util.ToJSON(b))

	if err := b.verify(); err != nil {
		response.HandleError(resp, errors.WithMessage(err, "backup name invalid"))
		return
	}

	if b.Location == "" || b.LocationConfig == nil {
		response.HandleError(resp, errors.New("backup location is required"))
		return
	}

	if b.BackupPolicies == nil || b.BackupPolicies.SnapshotFrequency == "" || b.BackupPolicies.TimesOfDay == "" {
		response.HandleError(resp, errors.New("backup policy is required"))
		return
	}

	var getLabel = "name=" + util.MD5(b.Name) + ",owner=" + owner
	backup, err := h.handler.GetBackupHandler().GetByLabel(ctx, getLabel)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.Errorf("failed to get backup %q: %v", b.Name, err))
		return
	}

	if backup != nil {
		response.HandleError(resp, errors.New("backup plan "+b.Name+" already exists"))
		return
	}

	createBackup, err := NewBackupPlan(owner, h.factory, h.handler).Apply(ctx, &b)
	if err != nil {
		response.HandleError(resp, errors.Errorf("failed to create backup %q: %v", b.Name, err))
		return
	}

	response.Success(resp, parseResponseBackupCreate(createBackup))
}

func (h *Handler) update(req *restful.Request, resp *restful.Response) {
	var (
		err error
		b   BackupCreate
	)

	if err = req.ReadEntity(&b); err != nil {
		response.HandleError(resp, err)
		return
	}

	backupId, owner := req.PathParameter("id"), req.HeaderParameter(constant.BflUserKey)
	ctx := req.Request.Context()
	b.Name = backupId

	log.Infof("received backup update request: %s", util.PrettyJSON(b))

	format := "failed to update backup plan %q"

	backup, err := h.handler.GetBackupHandler().GetById(ctx, backupId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessage(err, fmt.Sprintf("get backup %s error", backupId)))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("backup %s not found", backupId))
		return
	}

	if err = NewBackupPlan(owner, h.factory, h.handler).Update(ctx, &b, backup); err != nil {
		response.HandleError(resp, errors.WithMessagef(err, format, backupId))
		return
	}

	log.Infof("received backup update success, id: %s", backupId)

	backups, err := h.handler.GetBackupHandler().ListBackups(ctx, backupId, 0, 0, "", fmt.Sprintf("metadata.name=%s", backupId))
	if err != nil {
		response.HandleError(resp, errors.WithMessagef(err, format, backupId))
		return
	}

	result, _, totalPage := handlers.GenericPager(0, 0, backups)

	labelsSelector := h.handler.GetBackupHandler().GetBackupIdForLabels(result)
	var allSnapshots = new(sysv1.SnapshotList)
	for _, ls := range labelsSelector {
		snapshots, err := h.handler.GetSnapshotHandler().ListSnapshots(ctx, 0, 0, ls, "")
		if err != nil {
			log.Errorf("get snapshots error: %v", err)
			continue
		}
		if snapshots == nil || len(snapshots.Items) == 0 {
			continue
		}
		allSnapshots.Items = append(allSnapshots.Items, snapshots.Items...)
	}

	data := parseResponseBackupList(result, allSnapshots, totalPage)
	var items = data.Backups
	if len(items) == 0 {
		response.Success(resp, nil)
	} else {
		response.Success(resp, items[0])
	}
}

func (h *Handler) deleteBackupPlan(req *restful.Request, resp *restful.Response) {
	ctx, backupId := req.Request.Context(), req.PathParameter("id")

	log.Infof("delete backup: %s", backupId)

	backup, err := h.handler.GetBackupHandler().GetById(ctx, backupId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessagef(err, fmt.Sprintf("get backup %s error", backupId)))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("backup %s not found", backupId))
		return
	}

	if err := h.handler.GetBackupHandler().Delete(ctx, backup); err != nil {
		response.HandleError(resp, errors.WithMessagef(err, "delete backup %s error", backupId))
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) enabledBackupPlan(req *restful.Request, resp *restful.Response) {
	var (
		err error
		b   BackupEnabled
	)

	if err = req.ReadEntity(&b); err != nil {
		response.HandleError(resp, errors.WithStack(err))
		return
	}

	log.Infof("enabled backup plan request: %s", util.PrettyJSON(b))

	if !util.ListContains([]string{constant.BackupPause, constant.BackupResume}, strings.ToLower(b.Event)) {
		response.HandleError(resp, errors.WithMessagef(err, "backup event invalid, event: %s", b.Event))
		return
	}

	ctx, backupId := req.Request.Context(), req.PathParameter("id")

	log.Infof("backup: %s, event: %s", backupId, b.Event)

	backup, err := h.handler.GetBackupHandler().GetById(ctx, backupId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessagef(err, "get backup %s error", backupId))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("backup %s not found", backupId))
		return
	}

	if err := h.handler.GetBackupHandler().Enabled(ctx, backup, strings.ToLower(b.Event)); err != nil {
		response.HandleError(resp, errors.WithMessagef(err, "trigger backup %s error", backupId))
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) addSnapshot(req *restful.Request, resp *restful.Response) {
	var (
		err error
		b   CreateSnapshot
	)

	if err = req.ReadEntity(&b); err != nil {
		response.HandleError(resp, errors.WithStack(err))
		return
	}

	if b.Event != "create" {
		response.HandleError(resp, errors.Errorf("snapshot event invalid, event: %s", b.Event))
		return
	}

	ctx, backupId := req.Request.Context(), req.PathParameter("id")

	if backupId == "" {
		response.HandleError(resp, errors.New("backupId is empty"))
		return
	}

	backup, err := h.handler.GetBackupHandler().GetById(ctx, backupId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessagef(err, "get backup %s error", backupId))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("backup %s not found", backupId))
		return
	}

	if backup.Spec.Deleted {
		response.HandleError(resp, errors.WithMessagef(err, "backup %s is deleted", backupId))
		return
	}

	var location string
	for k := range backup.Spec.Location {
		location = k
		break
	}

	_, err = h.handler.GetSnapshotHandler().Create(ctx, backup, location)
	if err != nil {
		response.HandleError(resp, errors.WithMessagef(err, "create snapshot NOW error: %v, backupId: %s", err, backupId))
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) listSnapshots(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	backupId := req.PathParameter("id")
	limit := util.ParseToInt64(req.QueryParameter("limit"))
	offset := util.ParseToInt64(req.QueryParameter("offset"))

	var labelSelector = "backup-id=" + backupId
	var snapshots, err = h.handler.GetSnapshotHandler().ListSnapshots(ctx, offset, limit, labelSelector, "")
	if err != nil {
		response.HandleError(resp, errors.WithMessage(err, fmt.Sprintf("get backup %s snapshots error", backupId)))
		return
	}

	if snapshots == nil || len(snapshots.Items) == 0 {
		response.Success(resp, parseResponseSnapshotList(snapshots, 1, 1, 0))
		return
	}
	totalCount := int64(len(snapshots.Items))
	result, currentPage, totalPage := handlers.GenericPager(limit, offset, snapshots)

	response.Success(resp, parseResponseSnapshotList(result, currentPage, totalPage, totalCount))
}

func (h *Handler) getSnapshot(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	snapshotId := req.PathParameter("snapshotId")

	snapshot, err := h.handler.GetSnapshotHandler().GetById(ctx, snapshotId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessage(err, fmt.Sprintf("snapshot %s get error", snapshotId)))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("snapshot %s not found", snapshotId))
		return
	}

	response.Success(resp, parseResponseSnapshotDetail(snapshot))
}

func (h *Handler) getSnapshotOne(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	snapshotId := req.PathParameter("snapshotId")

	snapshot, err := h.handler.GetSnapshotHandler().GetById(ctx, snapshotId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessage(err, fmt.Sprintf("snapshot %s get error", snapshotId)))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("snapshot %s not found", snapshotId))
		return
	}

	response.Success(resp, parseResponseSnapshotOne(snapshot))
}

func (h *Handler) cancelSnapshot(req *restful.Request, resp *restful.Response) {
	var (
		err error
		b   SnapshotCancel
	)

	if err = req.ReadEntity(&b); err != nil {
		response.HandleError(resp, errors.WithStack(err))
		return
	}

	if b.Event != constant.BackupCancel {
		response.HandleError(resp, errors.WithMessagef(err, "snapshot event invalid, event: %s", b.Event))
		return
	}

	ctx := req.Request.Context()
	backupId := req.PathParameter("id")
	snapshotId := req.PathParameter("snapshotId")
	_ = ctx

	log.Debugf("backup: %s, snapshot: %s, event: %s", backupId, snapshotId, b.Event)

	_, err = h.handler.GetBackupHandler().GetById(ctx, backupId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessage(err, fmt.Sprintf("get backup %s error", backupId)))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("backup %s not found", backupId))
		return
	}

	snapshot, err := h.handler.GetSnapshotHandler().GetById(ctx, snapshotId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessagef(err, "get snapshot %s error", snapshotId))
		return
	}

	if snapshot == nil {
		response.HandleError(resp, fmt.Errorf("snapshot %s not exists", backupId))
		return
	}

	// Failed
	var phase = *snapshot.Spec.Phase
	if util.ListContains([]string{
		constant.Failed.String(), constant.Completed.String()}, phase) {
		log.Infof("snapshot %s phase %s no need to Cancel", snapshotId, phase)
		response.SuccessNoData(resp)
		return
	}

	if err := h.handler.GetSnapshotHandler().UpdatePhase(ctx, snapshotId, constant.Canceled.String(), constant.MessageTaskCanceled); err != nil {
		response.HandleError(resp, errors.WithMessagef(err, "update snapshot %s Canceled error", snapshotId))
		return
	}

	response.SuccessNoData(resp)
}

func (h *Handler) getNode(req *restful.Request, resp *restful.Response) {
	var node = os.Getenv("NODE_NAME")

	response.Success(resp, node)
}

func (h *Handler) getSpaceRegions(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	owner := req.HeaderParameter(constant.BflUserKey)

	olaresId, err := h.handler.GetSnapshotHandler().GetOlaresId(owner)
	if err != nil {
		response.HandleError(resp, errors.WithMessagef(err, "get olares id error"))
		return
	}

	var storageRegion = &storage.StorageRegion{
		Handlers: h.handler,
	}
	regions, err := storageRegion.GetRegions(ctx, owner, olaresId)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	response.Success(resp, regions)
}

func (h *Handler) listRestore(req *restful.Request, resp *restful.Response) {
	ctx := req.Request.Context()
	owner := req.HeaderParameter(constant.BflUserKey)
	limit := util.ParseToInt64(req.QueryParameter("limit"))
	offset := util.ParseToInt64(req.QueryParameter("offset"))

	restores, err := h.handler.GetRestoreHandler().ListRestores(ctx, owner, offset, limit)
	if err != nil {
		log.Errorf("get restores error: %v", err)
		response.Success(resp, parseResponseRestoreList(restores, 1))
		return
	}

	result, _, totalPage := handlers.GenericPager(limit, offset, restores)
	log.Debugf("list restore result length: %d", len(result.Items))

	response.Success(resp, parseResponseRestoreList(result, totalPage))
}

func (h *Handler) checkBackupUrl(req *restful.Request, resp *restful.Response) {
	var (
		err error
		b   RestoreCheckBackupUrl
	)

	if err = req.ReadEntity(&b); err != nil {
		log.Errorf("check backup url read entity error: %v", err)
		response.HandleError(resp, errors.Errorf("check backup url read entity error: %v", err))
		return
	}

	ctx := req.Request.Context()
	owner := req.HeaderParameter(constant.BflUserKey)

	log.Infof("received restore check backup url request: %s", util.ToJSON(b))

	urlInfo, err := handlers.ParseBackupUrl(owner, b.BackupUrl)
	if err != nil {
		log.Errorf("parse backup url error: %v", err)
		response.HandleError(resp, errors.New(constant.MessageBackupUrlIncorrect))
		return
	}

	if urlInfo.Location == constant.BackupLocationSpace.String() {
		response.HandleError(resp, errors.Errorf("backup location is space, no need to check url snapshots, url: %s", b.BackupUrl))
		return
	}

	log.Infof("format backup url info: %s", util.ToJSON(urlInfo))

	var storageSnapshots = &storage.StorageSnapshots{
		Handlers: h.handler,
	}
	snapshots, err := storageSnapshots.GetSnapshots(ctx, b.Password, owner, urlInfo.Location, urlInfo.Endpoint, urlInfo.BackupName, urlInfo.BackupId)
	if err != nil {
		log.Errorf("check backup url snapshots error: %v", err)
		response.HandleError(resp, err)
		return
	}

	if snapshots == nil || len(*snapshots) == 0 {
		response.HandleError(resp, errors.New(constant.MessageBackupUrlIncorrect))
		return
	}

	items, backupTypeTag, backupTypeAppName := h.handler.GetSnapshotHandler().SortSnapshotList(snapshots)

	result, totalCount, totalPage := handlers.GenericPager(b.Limit, b.Offset, items)

	response.Success(resp, parseCheckBackupUrl(result, urlInfo.BackupName, backupTypeTag, backupTypeAppName, urlInfo.Location, urlInfo.PvcPath, totalCount, totalPage))
}

func (h *Handler) addRestore(req *restful.Request, resp *restful.Response) {
	var (
		err error
		b   RestoreCreate
	)

	if err = req.ReadEntity(&b); err != nil {
		log.Errorf("add restore read entity error: %v", err)
		response.HandleError(resp, errors.WithStack(err))
		return
	}

	ctx := req.Request.Context()
	owner := req.HeaderParameter(constant.BflUserKey)

	log.Infof("received restore create request: %s", util.ToJSON(b))

	if err = b.verify(); err != nil {
		log.Errorf("add restore params invalid, params: %s, error: %v", util.ToJSON(b), err)
		response.HandleError(resp, err)
		return
	}

	clusterId, err := handlers.GetClusterId()
	if err != nil {
		response.HandleError(resp, errors.Errorf("get clusterId error: %v", err))
		return
	}

	var restoreTypeName = constant.RestoreTypeUrl
	var backupId, backupName, snapshotId, resticSnapshotId, snapshotTime, backupPath, location string
	var backupStorageInfo *handlers.RestoreBackupUrlDetail
	var urlInfo *handlers.BackupUrlType
	var getSnapshot *backupssdkrestic.Snapshot

	var backupType, backupAppName string

	if b.SnapshotId != "" {
		snapshotId = b.SnapshotId
		restoreTypeName = constant.RestoreTypeSnapshot
		snapshot, err := h.handler.GetSnapshotHandler().GetById(ctx, b.SnapshotId)
		if err != nil && !apierrors.IsNotFound(err) {
			response.HandleError(resp, errors.WithMessage(err, fmt.Sprintf("get snapshot %s error", snapshotId)))
			return
		}

		if apierrors.IsNotFound(err) {
			response.HandleError(resp, fmt.Errorf("snapshot %s not found", snapshotId))
			return
		}

		backup, err := h.handler.GetBackupHandler().GetById(ctx, snapshot.Spec.BackupId)
		if err != nil && !apierrors.IsNotFound(err) {
			response.HandleError(resp, errors.WithMessage(err, fmt.Sprintf("get backup %s error", snapshot.Spec.BackupId)))
			return
		}

		locationConfig, err := handlers.GetBackupLocationConfig(backup)
		if err != nil {
			response.HandleError(resp, errors.WithMessage(err, fmt.Sprintf("get backup location config %s error", snapshot.Spec.BackupId)))
			return
		}
		location = locationConfig["location"]

		if apierrors.IsNotFound(err) {
			response.HandleError(resp, fmt.Errorf("backup %s not found", snapshot.Spec.BackupId))
			return
		}

		backupId = backup.Name
		backupName = backup.Spec.Name
		backupPath = handlers.GetBackupPath(backup)
		resticSnapshotId = *snapshot.Spec.SnapshotId
		snapshotTime = fmt.Sprintf("%d", snapshot.Spec.CreateAt.Unix())
	} else {
		// ~ parse and split BackupURL
		u, err := util.Base64decode(b.BackupUrl)
		if err != nil || string(u) == "" {
			log.Errorf("parse BackupURL invalid, url: %s", b.BackupUrl)
			response.HandleError(resp, errors.New(constant.MessageBackupUrlIncorrect))
			return
		}

		var urlDecode = util.TrimLineBreak(string(u))

		urlInfo, err = handlers.ParseBackupUrl(owner, urlDecode)
		if err != nil {
			log.Errorf("parse BackupURL endpoint error: %v, url: %s", err, urlDecode)
			response.HandleError(resp, errors.New(constant.MessageBackupUrlIncorrect))
			return
		}
		log.Infof("urlInfo: %s", util.ToJSON(urlInfo))

		backupStorageInfo, backupName, backupId, resticSnapshotId, snapshotTime, location, err = handlers.ParseRestoreBackupUrlDetail(owner, urlDecode)
		if err != nil {
			log.Errorf("parse BackupURL detail error: %v, url: %s", err, urlDecode)
			response.HandleError(resp, errors.New(constant.MessageBackupUrlIncorrect))
			return
		}

		log.Infof("storageInfo: %s", util.ToJSON(backupStorageInfo))

		var storageSnapshots = &storage.StorageSnapshots{
			Handlers: h.handler,
		}

		// check snapshot exists
		getSnapshot, err = storageSnapshots.GetSnapshot(ctx, b.Password, owner, urlInfo.Location, urlInfo.Endpoint, urlInfo.BackupName, urlInfo.BackupId, resticSnapshotId, backupStorageInfo.CloudName, backupStorageInfo.RegionId, clusterId)
		if err != nil {
			log.Errorf("get snapshot error: %v", err)
			response.HandleError(resp, err)
			return
		}

		log.Infof("add restore, get snapshot data: %s", util.ToJSON(getSnapshot))

		backupType = handlers.GetBackupTypeFromTags(getSnapshot.Tags)
		if backupType == constant.BackupTypeApp {
			backupAppName = handlers.GetBackupTypeAppName(getSnapshot.Tags)
		} else {
			backupPath = handlers.GetBackupTypeFilePath(getSnapshot.Tags)
		}
	}

	if backupType == constant.BackupTypeFile {
		if b.Path == "" || b.SubPath == "" {
			response.HandleError(resp, fmt.Errorf("path or subPath is empty"))
			return
		}
	} else {
		if ok := h.handler.GetBackupHandler().CheckAppInstalled(fmt.Sprintf("wise-%s-wise", owner)); !ok {
			response.HandleError(resp, fmt.Errorf("Wise is not installed. Please go to the Market to install this app."))
			return
		}
	}

	var restoreType = &handlers.RestoreType{
		Owner:               owner,
		Type:                restoreTypeName,
		BackupId:            backupId,
		BackupName:          backupName,
		BackupUrl:           backupStorageInfo, // if snapshot,it will be nil
		Password:            util.Base64encode([]byte(strings.TrimSpace(b.Password))),
		SnapshotId:          snapshotId, // backupUrl is nil
		SnapshotTime:        snapshotTime,
		ResticSnapshotId:    resticSnapshotId,
		ClusterId:           clusterId, // backupUrl is nil
		Location:            location,
		Endpoint:            urlInfo.Endpoint,
		TotalBytesProcessed: getSnapshot.Summary.TotalBytesProcessed,
	}

	if backupType == constant.BackupTypeApp {
		restoreType.Name = backupAppName
	} else {
		restoreType.Path = b.Path
		restoreType.SubPath = strings.TrimSpace(b.SubPath)
		restoreType.SubPathTimestamp = time.Now().Unix()
		restoreType.BackupPath = backupPath
	}

	log.Infof("create restore task: %s", util.ToJSON(restoreType))

	restore, err := h.handler.GetRestoreHandler().CreateRestore(ctx, backupType, restoreType)
	if err != nil {
		response.HandleError(resp, errors.Errorf("create restore task failed: %v", err))
		return
	}

	response.Success(resp, parseResponseRestoreCreate(restore, backupName, snapshotTime, restoreType.Path))
}

func (h *Handler) getRestore(req *restful.Request, resp *restful.Response) {
	ctx, restoreId := req.Request.Context(), req.PathParameter("id")
	owner := req.HeaderParameter(constant.BflUserKey)
	_ = owner

	restore, err := h.handler.GetRestoreHandler().GetById(ctx, restoreId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessagef(err, "get restore %s error", restoreId))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("restore %s not found", restoreId))
		return
	}

	result, err := parseResponseRestoreDetailFromBackupUrl(restore)
	if err != nil {
		log.Errorf("get restore %s detail error: %s", restoreId, err)
	}
	response.Success(resp, result)
}

func (h *Handler) getRestoreOne(req *restful.Request, resp *restful.Response) {
	ctx, restoreId := req.Request.Context(), req.PathParameter("id")
	owner := req.HeaderParameter(constant.BflUserKey)
	_ = owner

	restore, err := h.handler.GetRestoreHandler().GetById(ctx, restoreId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessagef(err, "get restore %s error", restoreId))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("restore %s not found", restoreId))
		return
	}

	backupType, err := handlers.GetRestoreType(restore)
	if err != nil {
		response.HandleError(resp, errors.WithMessagef(err, "restore type %s invalid, error: %v", restoreId, err))
		return
	}

	var backupAppTypeName string
	if backupType == constant.BackupTypeApp {
		backupAppTypeName = handlers.GetRestoreAppName(restore)
	}

	restoreType, err := handlers.ParseRestoreType(backupType, restore)
	if err != nil {
		response.HandleError(resp, fmt.Errorf("parse %s restore type error: %v", restoreId, err))
		return
	}

	response.Success(resp, parseResponseRestoreOne(restore, backupAppTypeName, restoreType.BackupName, restoreType.SnapshotTime, restoreType.Path, restoreType.SubPath))
}

func (h *Handler) cancelRestore(req *restful.Request, resp *restful.Response) {
	var (
		err error
		b   RestoreCancel // support cancel, delete
	)

	if err = req.ReadEntity(&b); err != nil {
		response.HandleError(resp, errors.WithStack(err))
		return
	}

	if b.Event != constant.BackupCancel && b.Event != constant.BackupDelete {
		response.HandleError(resp, errors.WithMessagef(err, "restore event invalid, event: %s", b.Event))
		return
	}

	ctx := req.Request.Context()
	restoreId := req.PathParameter("id")

	log.Debugf("restore: %s, event: %s", restoreId, b.Event)

	restore, err := h.handler.GetRestoreHandler().GetById(ctx, restoreId)
	if err != nil && !apierrors.IsNotFound(err) {
		response.HandleError(resp, errors.WithMessagef(err, "get restore %s error", restoreId))
		return
	}

	if apierrors.IsNotFound(err) {
		response.HandleError(resp, fmt.Errorf("restore %s not found", restoreId))
		return
	}

	var phase = *restore.Spec.Phase
	if util.ListContains([]string{
		constant.Pending.String(), constant.Running.String()}, phase) {
		if err := h.handler.GetRestoreHandler().UpdatePhase(ctx, restoreId, constant.Canceled.String()); err != nil {
			response.HandleError(resp, errors.WithMessagef(err, "cancel restore %s error", restoreId))
			return
		}
	}

	if b.Event == constant.BackupDelete {
		if err := h.handler.GetRestoreHandler().UpdatePhase(ctx, restoreId, constant.Deleted.String()); err != nil {
			response.HandleError(resp, errors.WithMessagef(err, "delete restore %s error", restoreId))
			return
		}
	}

	response.SuccessNoData(resp)
}
