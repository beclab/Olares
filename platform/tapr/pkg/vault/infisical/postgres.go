package infisical

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"bytetrade.io/web3os/tapr/pkg/postgres"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"k8s.io/klog/v2"
)

type PostgresClient struct {
	DB *postgres.DBLogger
}

/*
	export const UsersSchema = z.object({
	  id: z.string().uuid(),
	  email: z.string().nullable().optional(),
	  authMethods: z.string().array().nullable().optional(),
	  superAdmin: z.boolean().default(false).nullable().optional(),
	  firstName: z.string().nullable().optional(),
	  lastName: z.string().nullable().optional(),
	  isAccepted: z.boolean().default(false).nullable().optional(),
	  isMfaEnabled: z.boolean().default(false).nullable().optional(),
	  mfaMethods: z.string().array().nullable().optional(),
	  devices: z.unknown().nullable().optional(),
	  createdAt: z.date(),
	  updatedAt: z.date(),
	  isGhost: z.boolean().default(false),
	  username: z.string(),
	  isEmailVerified: z.boolean().default(false).nullable().optional()
	});
*/
type UserPG struct {
	ID           *string  `db:"id,omitempty" json:"id,omitempty" mapstructure:"id,omitempty"`
	Email        string   `db:"email" json:"email" mapstructure:"email"`
	AuthMethods  []string `db:"authMethods,omitempty" json:"authMethods,omitempty" mapstructure:"authMethods,omitempty"`
	SuperAdmin   bool     `db:"superAdmin" json:"superAdmin" mapstructure:"superAdmin"`
	FirstName    string   `db:"firstName" json:"firstName" mapstructure:"firstName"`
	LastName     string   `db:"lastName" json:"lastName" mapstructure:"lastName"`
	IsAccepted   bool     `db:"isAccepted" json:"isAccepted" mapstructure:"isAccepted"`
	IsMfaEnabled bool     `db:"isMfaEnabled" json:"isMfaEnabled" mapstructure:"isMfaEnabled"`
	MfaMethods   []string `db:"mfaMethods,omitempty" json:"mfaMethods,omitempty" mapstructure:"mfaMethods,omitempty"`
	IsGhost      bool     `db:"isGhost" json:"isGhost" mapstructure:"isGhost"`
	Username     string   `db:"username" json:"username" mapstructure:"username"`
}

/*
	export const UserEncryptionKeysSchema = z.object({
	  id: z.string().uuid(),
	  clientPublicKey: z.string().nullable().optional(),
	  serverPrivateKey: z.string().nullable().optional(),
	  encryptionVersion: z.number().default(2).nullable().optional(),
	  protectedKey: z.string().nullable().optional(),
	  protectedKeyIV: z.string().nullable().optional(),
	  protectedKeyTag: z.string().nullable().optional(),
	  publicKey: z.string(),
	  encryptedPrivateKey: z.string(),
	  iv: z.string(),
	  tag: z.string(),
	  salt: z.string(),
	  verifier: z.string(),
	  userId: z.string().uuid()
	});
*/
type UserEncryptionKeysPG struct {
	ID                  *string `db:"id,omitempty" json:"id,omitempty" mapstructure:"id,omitempty"`
	ClientPublicKey     string  `db:"clientPublicKey" json:"clientPublicKey" mapstructure:"clientPublicKey"`
	ServerPrivateKey    string  `db:"serverPrivateKey" json:"serverPrivateKey" mapstructure:"serverPrivateKey"`
	EncryptionVersion   int32   `db:"encryptionVersion" json:"encryptionVersion" mapstructure:"encryptionVersion"`
	ProtectedKey        string  `db:"protectedKey" json:"protectedKey" mapstructure:"protectedKey"`
	ProtectedKeyIV      string  `db:"protectedKeyIV" json:"protectedKeyIV" mapstructure:"protectedKeyIV"`
	ProtectedKeyTag     string  `db:"protectedKeyTag" json:"protectedKeyTag" mapstructure:"protectedKeyTag"`
	PublicKey           string  `db:"publicKey" json:"publicKey" mapstructure:"publicKey"`
	EncryptedPrivateKey string  `db:"encryptedPrivateKey" json:"encryptedPrivateKey" mapstructure:"encryptedPrivateKey"`
	IV                  string  `db:"iv" json:"iv" mapstructure:"iv"`
	Tag                 string  `db:"tag" json:"tag" mapstructure:"tag"`
	Salt                string  `db:"salt" json:"salt" mapstructure:"salt"`
	Verifier            string  `db:"verifier" json:"verifier" mapstructure:"verifier"`
	UserID              string  `db:"userId" json:"userId" mapstructure:"userId"`
	OrgId               *string `db:"orgId,omitempty" json:"orgId,omitempty" mapstructure:"orgId,omitempty"`
}

/*
	export const OrganizationsSchema = z.object({
	  id: z.string().uuid(),
	  name: z.string(),
	  customerId: z.string().nullable().optional(),
	  slug: z.string(),
	  createdAt: z.date(),
	  updatedAt: z.date(),
	  authEnforced: z.boolean().default(false).nullable().optional(),
	  scimEnabled: z.boolean().default(false).nullable().optional()
	});
*/
type OrganizationsPG struct {
	ID           *string    `db:"id,omitempty" json:"id,omitempty" mapstructure:"id,omitempty"`
	Name         string     `db:"name" json:"name" mapstructure:"name"`
	CustomerId   string     `db:"customerId" json:"customerId" mapstructure:"customerId"`
	Slug         string     `db:"slug" json:"slug" mapstructure:"slug"`
	AuthEnforced bool       `db:"authEnforced" json:"authEnforced" mapstructure:"authEnforced"`
	ScimEnabled  bool       `db:"scimEnabled" json:"scimEnabled" mapstructure:"scimEnabled"`
	CreatedAt    *time.Time `db:"createdAt,omitempty" json:"createdAt,omitempty" mapstructure:"createdAt,omitempty"`
	UpdatedAt    *time.Time `db:"updatedAt,omitempty" json:"updatedAt,omitempty" mapstructure:"updatedAt,omitempty"`
}

/*
	export const OrgMembershipsSchema = z.object({
	  id: z.string().uuid(),
	  role: z.string(),
	  status: z.string().default("invited"),
	  inviteEmail: z.string().nullable().optional(),
	  createdAt: z.date(),
	  updatedAt: z.date(),
	  userId: z.string().uuid().nullable().optional(),
	  orgId: z.string().uuid(),
	  roleId: z.string().uuid().nullable().optional()
	});
*/
type OrgMembershipsPG struct {
	ID          *string `db:"id,omitempty" json:"id,omitempty" mapstructure:"id,omitempty"`
	Role        string  `db:"role" json:"role" mapstructure:"role"`
	Status      string  `db:"status" json:"status" mapstructure:"status"`
	InviteEmail string  `db:"inviteEmail" json:"inviteEmail" mapstructure:"inviteEmail"`
	UserId      string  `db:"userId" json:"userId" mapstructure:"userId"`
	OrgId       string  `db:"orgId" json:"orgId" mapstructure:"orgId"`
	RoleId      *string `db:"roleId,omitempty" json:"roleId,omitempty" mapstructure:"roleId,omitempty"`
}

/*
	export const AuthTokenSessionsSchema = z.object({
		id: z.string().uuid(),
		ip: z.string(),
		userAgent: z.string().nullable().optional(),
		refreshVersion: z.number().default(1),
		accessVersion: z.number().default(1),
		lastUsed: z.date(),
		createdAt: z.date(),
		updatedAt: z.date(),
		userId: z.string().uuid()
	  });
*/
type AuthTokenSessionsPG struct {
	ID             *string    `db:"id,omitempty" json:"id,omitempty" mapstructure:"id,omitempty"`
	IP             string     `db:"ip" json:"ip" mapstructure:"ip"`
	UserAgent      *string    `db:"userAgent,omitempty" json:"userAgent,omitempty" mapstructure:"userAgent,omitempty"`
	RefreshVersion int        `db:"refreshVersion" json:"refreshVersion" mapstructure:"refreshVersion"`
	AccessVersion  int        `db:"accessVersion" json:"accessVersion" mapstructure:"accessVersion"`
	UserId         string     `db:"userId" json:"userId" mapstructure:"userId"`
	LastUsed       time.Time  `db:"lastUsed" json:"lastUsed" mapstructure:"lastUsed"`
	CreatedAt      *time.Time `db:"createdAt,omitempty" json:"createdAt,omitempty" mapstructure:"createdAt,omitempty"`
	UpdatedAt      *time.Time `db:"updatedAt,omitempty" json:"updatedAt,omitempty" mapstructure:"updatedAt,omitempty"`
}

func (c *PostgresClient) Close() {
	err := c.DB.Close()
	if err != nil {
		klog.Error("close db error, ", err)
	}
}

func NewClient(dsn string) (*PostgresClient, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	dbProxy := postgres.DBLogger{DB: db}

	dbProxy.Debug()

	return &PostgresClient{DB: &dbProxy}, nil
}

func (c *PostgresClient) UpdateSuperAdmin(basectx context.Context) (orgId string, err error) {
	ctx, cancel := context.WithTimeout(basectx, 10*time.Second)
	defer cancel()

	sql := "update super_admin set initialized = true"
	_, err = c.DB.ExecContext(ctx, sql)
	if err != nil {
		klog.Error("update super_admin error, ", err)
		return "", err
	}

	// create or get org id
	sql = "select * from organizations where name = :name"
	res, err := c.DB.NamedQueryContext(ctx, sql, map[string]interface{}{
		"name": "Terminus",
	})
	if err != nil {
		klog.Error("fetch org error,  ", err)

		return "", err
	}

	var org OrganizationsPG
	if res.Next() {
		err = res.StructScan(&org)
		if err != nil {
			klog.Error("scan org data error, ", err)
			return "", err
		}
		return *org.ID, nil
	}

	org = OrganizationsPG{
		Name: "Terminus",
	}

	orgId, err = org.Create(basectx, c)
	if err != nil {
		klog.Error("create org error,  ", err)
	}

	return orgId, nil
}

func (c *PostgresClient) GetUser(basectx context.Context, email string) (*UserEncryptionKeysPG, error) {
	if email == "" {
		return nil, errors.New("email is empty")
	}

	ctx, cancel := context.WithTimeout(basectx, 10*time.Second)
	defer cancel()

	sql := "select b.*, c.\"orgId\" from users a, user_encryption_keys b, org_memberships c where a.email=:email and a.id = b.\"userId\" and a.id = c.\"userId\""
	res, err := c.DB.NamedQueryContext(ctx, sql, map[string]interface{}{
		"email": email,
	})

	if err != nil {
		klog.Error("fetch user error, ", err)

		return nil, err
	}

	var user UserEncryptionKeysPG
	if res.Next() {
		err = res.StructScan(&user)
		if err != nil {
			klog.Error("scan user data error, ", err)
			return nil, err
		}
		return &user, nil
	}

	return nil, nil
}

func (c *PostgresClient) GetUserTokenSession(basectx context.Context, userId, ip, userAgent string) (*AuthTokenSessionsPG, error) {
	if userId == "" {
		return nil, errors.New("user is empty")
	}

	ctx, cancel := context.WithTimeout(basectx, 10*time.Second)
	defer cancel()

	sql := "select * from auth_token_sessions where \"userId\"=:userId and ip = :ip and \"userAgent\" = :userAgent"
	res, err := c.DB.NamedQueryContext(ctx, sql, map[string]interface{}{
		"userId":    userId,
		"ip":        ip,
		"userAgent": userAgent,
	})

	if err != nil {
		klog.Error("fetch auth token session error, ", err)

		return nil, err
	}

	var session AuthTokenSessionsPG
	if res.Next() {
		err = res.StructScan(&session)
		if err != nil {
			klog.Error("scan token session data error, ", err)
			return nil, err
		}
		return &session, nil
	}

	session.UserId = userId
	session.IP = ip
	session.UserAgent = &userAgent
	session.AccessVersion = 1
	session.RefreshVersion = 1
	session.LastUsed = time.Now()

	sid, err := session.Create(ctx, c)
	if err != nil {
		return nil, err
	}

	session.ID = &sid

	return &session, nil
}

func (c *PostgresClient) SaveUser(basectx context.Context, orgId string, user *UserPG, userEnc *UserEncryptionKeysPG) (string, error) {
	if user == nil {
		return "", errors.New("user is empty")
	}

	uid, err := user.Create(basectx, c)
	if err != nil {
		return "", err
	}

	if userEnc != nil {
		userEnc.UserID = uid
		_, err := userEnc.Create(basectx, c)
		if err != nil {
			return "", err
		}
	}

	member := OrgMembershipsPG{
		OrgId:  orgId,
		UserId: uid,
		Role:   "admin",
		Status: "accepted",
	}

	_, err = member.Create(basectx, c)
	if err != nil {
		return "", err
	}

	return uid, nil
}

func (c *PostgresClient) DeleteUser(basectx context.Context, userId string) error {
	err := delete(basectx, c, "user", fmt.Sprintf("id = %s", userId))
	if err != nil {
		klog.Error("delete user error, ", err)
		return err
	}

	err = delete(basectx, c, "user_encryption_keys", fmt.Sprintf("userId = %s", userId))
	if err != nil {
		klog.Error("delete user encryption keys error, ", err)
	}

	err = delete(basectx, c, "org_memberships", fmt.Sprintf("userId = %s", userId))
	if err != nil {
		klog.Error("delete org memberships error, ", err)
	}

	return nil
}

func ValueMapper[T interface{}](obj T) (fields, namedKeys []string, err error) {
	values := make(map[string]interface{})
	err = mapstructure.Decode(obj, &values)
	if err != nil {
		klog.Error("decode object value error, ", err)
		return
	}

	fields = make([]string, 0, len(values))
	namedKeys = make([]string, 0, len(values))
	for k := range values {
		fields = append(fields, "\""+k+"\"")
		namedKeys = append(namedKeys, ":"+k)
	}

	return
}

func insert[T interface{}](basectx context.Context, client *PostgresClient, table string, obj T, setId func(T, string) T) (id string, err error) {
	id = uuid.New().String()

	obj = setId(obj, id)

	fields, keys, err := ValueMapper(obj)
	if err != nil {
		return
	}

	sql := fmt.Sprintf("insert into %s(%s) values(%s)", table, strings.Join(fields, ","), strings.Join(keys, ","))
	ctx, cancel := context.WithTimeout(basectx, 10*time.Second)
	defer cancel()

	_, err = client.DB.NamedExecContext(ctx, sql, obj)
	if err != nil {
		klog.Error("create error, ", err, ", ", table)
		return
	}

	return

}

func delete(basectx context.Context, client *PostgresClient, table string, whereClause string) error {
	sql := fmt.Sprintf("delete from %s %s", table, whereClause)

	ctx, cancel := context.WithTimeout(basectx, 10*time.Second)
	defer cancel()

	_, err := client.DB.ExecContext(ctx, sql)
	if err != nil {
		klog.Error("delete error, ", err, ", ", table)
		return err
	}

	return nil
}

func (u *UserPG) Create(basectx context.Context, client *PostgresClient) (id string, err error) {
	return insert(basectx, client, "users", u, func(obj *UserPG, id string) *UserPG {
		obj.ID = &id
		return obj
	})
}

func (u *UserEncryptionKeysPG) Create(basectx context.Context, client *PostgresClient) (id string, err error) {
	return insert(basectx, client, "user_encryption_keys", u, func(obj *UserEncryptionKeysPG, id string) *UserEncryptionKeysPG {
		obj.ID = &id
		return obj
	})
}

func (o *OrganizationsPG) Create(basectx context.Context, client *PostgresClient) (id string, err error) {
	return insert(basectx, client, "organizations", o, func(obj *OrganizationsPG, id string) *OrganizationsPG {
		obj.ID = &id
		return obj
	})
}

func (o *OrgMembershipsPG) Create(basectx context.Context, client *PostgresClient) (id string, err error) {
	return insert(basectx, client, "org_memberships", o, func(obj *OrgMembershipsPG, id string) *OrgMembershipsPG {
		obj.ID = &id
		return obj
	})
}

func (o *AuthTokenSessionsPG) Create(basectx context.Context, client *PostgresClient) (id string, err error) {
	return insert(basectx, client, "auth_token_sessions", o, func(obj *AuthTokenSessionsPG, id string) *AuthTokenSessionsPG {
		obj.ID = &id
		return obj
	})
}
