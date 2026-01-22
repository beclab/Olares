package app

import (
	"fmt"
	"strconv"

	aprv1 "bytetrade.io/web3os/tapr/pkg/apis/apr/v1alpha1"
	"bytetrade.io/web3os/tapr/pkg/constants"
	"bytetrade.io/web3os/tapr/pkg/workload/citus"
	"bytetrade.io/web3os/tapr/pkg/workload/clickhouse"
	"bytetrade.io/web3os/tapr/pkg/workload/elasticsearch"
	"bytetrade.io/web3os/tapr/pkg/workload/mariadb"
	"bytetrade.io/web3os/tapr/pkg/workload/minio"
	"bytetrade.io/web3os/tapr/pkg/workload/mongodb"
	wmysql "bytetrade.io/web3os/tapr/pkg/workload/mysql"
	"bytetrade.io/web3os/tapr/pkg/workload/nats"
	"bytetrade.io/web3os/tapr/pkg/workload/rabbitmq"
	rediscluster "bytetrade.io/web3os/tapr/pkg/workload/redis-cluster"
	"bytetrade.io/web3os/tapr/pkg/workload/zinc"

	"github.com/gofiber/fiber/v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
)

func (s *Server) handleGetMiddlewareRequestInfo(ctx *fiber.Ctx) error {
	mwReq := &MiddlewareReq{}
	err := ctx.BodyParser(mwReq)
	if err != nil {
		klog.Error("parse request body error, ", err, ", ", string(ctx.Body()))
		return err
	}

	middlewares, err := s.MrLister.MiddlewareRequests(mwReq.Namespace).List(labels.Everything())
	if err != nil {
		klog.Error("get middleware list error, ", err)
		return err
	}

	for _, m := range middlewares {
		if m.Spec.App == mwReq.App &&
			m.Spec.AppNamespace == mwReq.AppNamespace &&
			m.Spec.Middleware == mwReq.Middleware {
			klog.Info("find middleware request cr")
			resp, err := s.getMiddlewareInfo(ctx, mwReq, m)
			if err != nil {
				return err
			}

			ctx.JSON(fiber.Map{
				"code": fiber.StatusOK,
				"data": resp,
			})

			return nil
		} // end of find middleware
	} // end of middleware loop

	return fiber.NewError(fiber.StatusNotFound, "middleware not found")
}

func (s *Server) handleListMiddlewareRequests(ctx *fiber.Ctx) error {
	middlewares, err := s.MrLister.List(labels.Everything())
	if err != nil {
		klog.Error("get middleware list error, ", err)
		return err
	}

	queryType := ctx.Query("middleware")
	var filterType aprv1.MiddlewareType
	if queryType != "" {
		filterType = aprv1.MiddlewareType(queryType)
	}

	var infos []*MiddlewareRequestInfo
	for _, m := range middlewares {
		// Optional filter by middleware type via query parameter
		if filterType != "" && m.Spec.Middleware != filterType {
			continue
		}
		var (
			user, pwd string
			err       error
			dbs       []Database
		)
		switch m.Spec.Middleware {
		case aprv1.TypeMongoDB:
			user = m.Spec.MongoDB.User
			pwd, err = m.Spec.MongoDB.Password.GetVarValue(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				klog.Error("get middleware mongo request password error, ", err)
				return err
			}

			for _, d := range m.Spec.MongoDB.Databases {
				dbs = append(dbs, Database{Name: d.Name})
			}

		case aprv1.TypePostgreSQL:
			user = m.Spec.PostgreSQL.User
			pwd, err = m.Spec.PostgreSQL.Password.GetVarValue(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				klog.Error("get middleware postgres request password error, ", err)
				return err
			}

			for _, d := range m.Spec.PostgreSQL.Databases {
				dbs = append(dbs, Database{Name: d.Name, Distributed: d.IsDistributed()})
			}

		case aprv1.TypeRedis:
			pwd, err = m.Spec.Redis.Password.GetVarValue(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				klog.Error("get middleware redis request password error, ", err)
				return err
			}

			dbs = append(dbs, Database{Name: m.Spec.Redis.Namespace})

		case aprv1.TypeZinc:
			user = m.Spec.Zinc.User
			pwd, err = m.Spec.Zinc.Password.GetVarValue(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				klog.Error("get middleware zinc request password error, ", err)
				return err
			}

			for _, idx := range m.Spec.Zinc.Indexes {
				dbs = append(dbs, Database{Name: zinc.GetIndexName(m.Spec.AppNamespace, idx.Name)})
			}

		case aprv1.TypeMinio:
			user = m.Spec.Minio.User
			pwd, err = m.Spec.Minio.Password.GetVarValue(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				klog.Error("get middleware minio request password error, ", err)
				return err
			}
			for _, b := range m.Spec.Minio.Buckets {
				dbs = append(dbs, Database{Name: minio.GetBucketName(m.Spec.AppNamespace, b.Name)})
			}
		case aprv1.TypeRabbitMQ:
			user = m.Spec.RabbitMQ.User
			pwd, err = m.Spec.RabbitMQ.Password.GetVarValue(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				klog.Error("get middleware rabbitmq request password error, ", err)
				return err
			}
			for _, b := range m.Spec.RabbitMQ.Vhosts {
				dbs = append(dbs, Database{Name: rabbitmq.GetVhostName(m.Spec.AppNamespace, b.Name)})
			}
		case aprv1.TypeElasticsearch:
			user = m.Spec.Elasticsearch.User
			pwd, err = m.Spec.Elasticsearch.Password.GetVarValue(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				klog.Error("get middleware elasticsearch request password error, ", err)
				return err
			}
			for _, b := range m.Spec.Elasticsearch.Indexes {
				dbs = append(dbs, Database{Name: elasticsearch.GetIndexName(m.Spec.AppNamespace, b.Name)})
			}
		case aprv1.TypeNats:
			user = m.Spec.Nats.User
			klog.Infof("nats: m.Namespace: %s", m.Namespace)
			pwd, err = m.Spec.Nats.Password.GetVarValue(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				klog.Errorf("get middleware nats request password error %v", err)
				return err
			}
			for _, s := range m.Spec.Nats.Subjects {
				dbs = append(dbs, Database{Name: fmt.Sprintf("%s.%s", m.Spec.AppNamespace, s.Name)})
			}

		}
		info := &MiddlewareRequestInfo{
			MetaInfo: MetaInfo{
				Name:      m.Name,
				Namespace: m.Namespace,
			},
			App: MetaInfo{
				Name:      m.Spec.App,
				Namespace: m.Spec.AppNamespace,
			},
			UserName:  user,
			Password:  pwd,
			Databases: dbs,
			//Buckets:   buckets,
			//Vhosts:    vhosts,
			//Indexes:   indexes,
			//Subjects:  subjects,
			Type: m.Spec.Middleware,
		}

		infos = append(infos, info)
	}

	return ctx.JSON(map[string]interface{}{
		"code": fiber.StatusOK,
		"data": infos,
	})
}

func (s *Server) handleListMiddlewares(ctx *fiber.Ctx) error {
	middleware := ctx.Params("middleware")
	var clusterResp []*MiddlewareClusterResp
	switch middleware {
	case string(aprv1.TypeRedis):
		klog.Info("list redis cluster crd")
		drcs, err := rediscluster.ListKvRocks(s.RedixLister)
		if err != nil {
			return err
		}

		for _, drc := range drcs {
			klog.Info("find redis cluster password")
			pwd, err := rediscluster.FindRedisClusterPassword(ctx.UserContext(), s.k8sClientSet, drc.Namespace)
			if err != nil {
				return err
			}

			cres := MiddlewareClusterResp{
				MiddlewareType: TypeRedis,
				MetaInfo: MetaInfo{
					Name:      drc.Name,
					Namespace: drc.Namespace,
				},
				Password: pwd,
				RedisProxy: Proxy{
					Endpoint: rediscluster.RedisClusterService + "." + drc.Namespace + ":" + strconv.Itoa(int(6379)),
				},
			}

			clusterResp = append(clusterResp, &cres)
		}

	case string(aprv1.TypeMongoDB):
		klog.Info("list percona mongo cluster crd")
		mdbs, err := mongodb.ListMongoClusters(ctx.UserContext(), s.ctrlClient, "")
		if err != nil {
			return err
		}

		for _, mdb := range mdbs {
			klog.Info("find mongo cluster password")
			user, pwd, err := mongodb.FindMongoAdminUser(ctx.UserContext(), s.k8sClientSet, "mongodb-middleware")
			if err != nil {
				return err
			}

			cres := MiddlewareClusterResp{
				MiddlewareType: TypeMongoDB,
				MetaInfo: MetaInfo{
					Name:      mdb.Name,
					Namespace: mdb.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Nodes:     1,
				Mongos: Proxy{
					Endpoint: mdb.Name + "-mongodb-headless." + mdb.Namespace + ":" + "27017",
					Size:     1,
				},
			}

			clusterResp = append(clusterResp, &cres)
		}

	case string(aprv1.TypePostgreSQL):
		klog.Info("list pg cluster crd")
		pgcs, err := s.PgLister.List(labels.Everything())
		if err != nil {
			klog.Error("list pg cluster error, ", err)
			return err
		}

		for _, pgc := range pgcs {
			klog.Info("find pg cluster password")
			user, pwd, err := citus.GetPGClusterAdminUserAndPassword(ctx.UserContext(), s.aprClientSet, s.k8sClientSet, pgc.Namespace)
			if err != nil {
				klog.Error("find pg cluster password error, ", err)
				return err
			}

			cres := MiddlewareClusterResp{
				MiddlewareType: TypePostgreSQL,
				MetaInfo: MetaInfo{
					Name:      pgc.Name,
					Namespace: pgc.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Nodes:     pgc.Spec.Replicas,
			}

			clusterResp = append(clusterResp, &cres)
		}
	case string(aprv1.TypeMinio):
		klog.Info("list minio cluster crd")
		minios, err := minio.ListMinioClusters(ctx.UserContext(), s.ctrlClient, "")
		if err != nil {
			return err
		}
		for _, m := range minios {
			user, pwd, err := minio.FindMinioAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeMinio,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: m.Name + "-minio-headless." + m.Namespace + ":" + "9000",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	case string(aprv1.TypeRabbitMQ):
		klog.Info("list rabbitmq cluster crd")
		rabbitmqs, err := rabbitmq.ListRabbitMQClusters(ctx.UserContext(), s.ctrlClient, "")
		if err != nil {
			return err
		}
		for _, m := range rabbitmqs {
			user, pwd, err := rabbitmq.FindRabbitMQAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeRabbitMQ,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: m.Name + "-rabbitmq-headless." + m.Namespace + ":" + "5672",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	case string(aprv1.TypeElasticsearch):
		klog.Info("list elasticsearch cluster crd")
		rabbitmqs, err := rabbitmq.ListRabbitMQClusters(ctx.UserContext(), s.ctrlClient, "")
		if err != nil {
			return err
		}
		for _, m := range rabbitmqs {
			user, pwd, err := elasticsearch.FindElasticsearchAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeElasticsearch,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: m.Name + "-mdit-http." + m.Namespace + ":" + "9200",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	case string(aprv1.TypeMariaDB):
		klog.Info("list mariadb cluster crd")
		mdbs, err := mariadb.ListMariadbClusters(ctx.UserContext(), s.ctrlClient, "")
		if err != nil {
			return err
		}
		for _, m := range mdbs {
			user, pwd, err := mariadb.FindMariaDBAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeMariaDB,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: m.Name + "-mariadb-headless." + m.Namespace + ":" + "3306",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	case string(aprv1.TypeMysql):
		klog.Info("list mysql cluster crd")
		mdbs, err := wmysql.ListMysqlClusters(ctx.UserContext(), s.ctrlClient, "")
		if err != nil {
			return err
		}
		for _, m := range mdbs {
			user, pwd, err := wmysql.FindMysqlAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeMySQL,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: m.Name + "-mysql-headless." + m.Namespace + ":" + "3306",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	case string(aprv1.TypeClickHouse):
		dbs, err := clickhouse.ListClickHouseClusters(ctx.UserContext(), s.ctrlClient, "")
		if err != nil {
			return err
		}
		for _, m := range dbs {
			user, pwd, err := clickhouse.FindClickHouseAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeClickHouse,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: "clickhouse-svc." + m.Namespace + ":" + "9000",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	default:
		return fiber.ErrNotFound
	}

	return ctx.JSON(map[string]interface{}{
		"code": fiber.StatusOK,
		"data": clusterResp,
	})
}
func (s *Server) handleListMiddlewaresAll(ctx *fiber.Ctx) error {
	var clusterResp []*MiddlewareClusterResp
	username := ctx.Context().UserValueBytes(constants.UsernameCtxKey)

	// Redis
	klog.Info("list redis cluster crd")
	if drcs, err := rediscluster.ListKvRocks(s.RedixLister); err == nil {
		for _, drc := range drcs {
			pwd, err := rediscluster.FindRedisClusterPassword(ctx.UserContext(), s.k8sClientSet, drc.Namespace)
			if err != nil {
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeRedis,
				MetaInfo: MetaInfo{
					Name:      drc.Name,
					Namespace: drc.Namespace,
				},
				Password: pwd,
				Proxy: Proxy{
					Endpoint: fmt.Sprintf("%s.%s:6379", rediscluster.RedisClusterService, fmt.Sprintf("user-system-%s", username.(string))),
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	} else {
		return err
	}

	// MongoDB
	klog.Info("list percona mongo cluster crd")
	if mdbs, err := mongodb.ListMongoClusters(ctx.UserContext(), s.ctrlClient, ""); err == nil {
		for _, mdb := range mdbs {
			user, pwd, err := mongodb.FindMongoAdminUser(ctx.UserContext(), s.k8sClientSet, "mongodb-middleware")
			if err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeMongoDB,
				MetaInfo: MetaInfo{
					Name:      mdb.Name,
					Namespace: mdb.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Nodes:     1,
				Proxy: Proxy{
					Endpoint: mdb.Name + "-mongodb-headless." + mdb.Namespace + ":" + "27017",
					Size:     1,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	} else {
		return err
	}

	// PostgreSQL
	klog.Info("list pg cluster crd")
	if pgcs, err := s.PgLister.List(labels.Everything()); err == nil {
		for _, pgc := range pgcs {
			user, pwd, err := citus.GetPGClusterAdminUserAndPassword(ctx.UserContext(), s.aprClientSet, s.k8sClientSet, pgc.Namespace)
			if err != nil {
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypePostgreSQL,
				MetaInfo: MetaInfo{
					Name:      pgc.Name,
					Namespace: pgc.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Nodes:     pgc.Spec.Replicas,
				Proxy: Proxy{
					Endpoint: fmt.Sprintf("%s.%s:5432", "citus-master-svc", fmt.Sprintf("user-system-%s", username.(string))),
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	} else {
		return err
	}

	// Minio
	klog.Info("list minio cluster crd")
	if minios, err := minio.ListMinioClusters(ctx.UserContext(), s.ctrlClient, ""); err == nil {
		for _, m := range minios {
			user, pwd, err := minio.FindMinioAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeMinio,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: m.Name + "-minio-headless." + m.Namespace + ":" + "9000",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	} else {
		return err
	}

	// RabbitMQ
	klog.Info("list rabbitmq cluster crd")
	if rabbitmqs, err := rabbitmq.ListRabbitMQClusters(ctx.UserContext(), s.ctrlClient, ""); err == nil {
		for _, m := range rabbitmqs {
			user, pwd, err := rabbitmq.FindRabbitMQAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeRabbitMQ,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: m.Name + "-rabbitmq-headless." + m.Namespace + ":" + "5672",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	} else {
		return err
	}

	// Elasticsearch
	klog.Info("list elasticsearch cluster crd")
	if elss, err := elasticsearch.ListRabbitMQClusters(ctx.UserContext(), s.ctrlClient, ""); err == nil {
		for _, m := range elss {
			user, pwd, err := elasticsearch.FindElasticsearchAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeElasticsearch,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: m.Name + "-mdit-http." + m.Namespace + ":" + "9200",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	} else {
		return err
	}

	// MariaDB
	klog.Info("list mariadb cluster crd")
	if mdbs, err := mariadb.ListMariadbClusters(ctx.UserContext(), s.ctrlClient, ""); err == nil {
		for _, m := range mdbs {
			user, pwd, err := mariadb.FindMariaDBAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeMariaDB,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: m.Name + "-mariadb-headless." + m.Namespace + ":" + "3306",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	} else {
		return err
	}

	// MySQL
	klog.Info("list mysql cluster crd")
	if mdbs, err := wmysql.ListMysqlClusters(ctx.UserContext(), s.ctrlClient, ""); err == nil {
		for _, m := range mdbs {
			user, pwd, err := wmysql.FindMysqlAdminUser(ctx.UserContext(), s.k8sClientSet, m.Namespace)
			if err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return err
			}
			cres := MiddlewareClusterResp{
				MiddlewareType: TypeMySQL,
				MetaInfo: MetaInfo{
					Name:      m.Name,
					Namespace: m.Namespace,
				},
				AdminUser: user,
				Password:  pwd,
				Proxy: Proxy{
					Endpoint: m.Name + "-mysql-headless." + m.Namespace + ":" + "3306",
					Size:     m.Spec.ComponentSpecs[0].Replicas,
				},
			}
			clusterResp = append(clusterResp, &cres)
		}
	} else {
		return err
	}
	user, pwd, err := nats.FindNatsAdminUser(ctx.UserContext(), s.k8sClientSet)
	if err != nil {
		return err
	}

	natsResp := MiddlewareClusterResp{
		MiddlewareType: TypeNats,
		MetaInfo: MetaInfo{
			Name:      "nats",
			Namespace: constants.PlatformNamespace,
		},
		AdminUser: user,
		Password:  pwd,
		Proxy: Proxy{
			Endpoint: fmt.Sprintf("%s.%s:%d", "nats", constants.PlatformNamespace, 4222),
		},
	}
	clusterResp = append(clusterResp, &natsResp)

	return ctx.JSON(map[string]interface{}{
		"code": fiber.StatusOK,
		"data": clusterResp,
	})
}

func (s *Server) handleScaleMiddleware(ctx *fiber.Ctx) error {
	scaleReq := ClusterScaleReq{}
	err := ctx.BodyParser(&scaleReq)
	if err != nil {
		klog.Error("parse request body error, ", err, ", ", string(ctx.Body()))
		return err
	}

	switch scaleReq.Middleware {
	case aprv1.TypeMongoDB:
		err = mongodb.ScalePerconaMongoNodes(ctx.UserContext(), s.dynamicClient, scaleReq.Name, scaleReq.Namespace, scaleReq.Nodes)
		if err != nil {
			return err
		}
	case aprv1.TypeRedis:
		err = rediscluster.ScaleRedisClusterNodes(ctx.UserContext(), s.dynamicClient, scaleReq.Name, scaleReq.Namespace, scaleReq.Nodes)
		if err != nil {
			return err
		}
	case aprv1.TypePostgreSQL:
		pgc, err := s.aprClientSet.AprV1alpha1().PGClusters(scaleReq.Namespace).Get(ctx.UserContext(), scaleReq.Name, metav1.GetOptions{})
		if err != nil {
			klog.Error("get current pg cluster to scale up error, ", err)
			return err
		}

		if pgc.Spec.Replicas > scaleReq.Nodes {
			klog.Error("scale down pg cluster is not implemented")
			return fiber.ErrNotImplemented
		}

		pgc.Spec.Replicas = scaleReq.Nodes

		if _, err = s.aprClientSet.AprV1alpha1().PGClusters(scaleReq.Namespace).
			Update(ctx.UserContext(), pgc, metav1.UpdateOptions{}); err != nil {
			klog.Error("update pg cluster replicas error, ", err)
			return err
		}

	default:
		return fiber.ErrNotImplemented
	}

	return ctx.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "scale success",
	})
}

func (s *Server) handleUpdateMiddlewareAdminPassword(ctx *fiber.Ctx) error {
	changePwdReq := ClusterChangePwdReq{}
	err := ctx.BodyParser(&changePwdReq)
	if err != nil {
		klog.Error("parse request body error, ", err, ", ", string(ctx.Body()))
		return err
	}

	if changePwdReq.Password == "" {
		klog.Error("password is empty")
		return fiber.ErrNotAcceptable
	}

	switch changePwdReq.Middleware {
	case aprv1.TypePostgreSQL:
		pgc, err := s.aprClientSet.AprV1alpha1().PGClusters(changePwdReq.Namespace).Get(ctx.UserContext(), changePwdReq.Name, metav1.GetOptions{})
		if err != nil {
			klog.Error("get current pg cluster to scale up error, ", err)
			return err
		}

		if changePwdReq.User != "" {
			pgc.Spec.AdminUser = changePwdReq.User
		}

		pgc.Spec.Password.Value = changePwdReq.Password
		pgc.Spec.Password.ValueFrom = nil

		_, err = s.aprClientSet.AprV1alpha1().PGClusters(changePwdReq.Namespace).Update(ctx.UserContext(), pgc, metav1.UpdateOptions{})
		if err != nil {
			klog.Error("update pg cluster error, ", err, ", ", changePwdReq.Name, ", ", changePwdReq.Namespace)
			return err
		}

	default:
		return fiber.ErrNotImplemented
	}

	return ctx.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "update success",
	})
}
