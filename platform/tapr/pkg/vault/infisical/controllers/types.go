package controllers

import "strings"

type Secret struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Environment string `json:"env"`
}

type RequiredOp struct {
	Op     string
	Params map[string]string
}

func DecodeOps(op string) *RequiredOp {
	opTokenizer := strings.Split(op, "?")

	rop := &RequiredOp{
		Op:     opTokenizer[0],
		Params: make(map[string]string),
	}

	if len(opTokenizer) > 1 {
		params := strings.Split(opTokenizer[1], "&")
		for _, p := range params {
			if p != "" {
				kv := strings.Split(p, "=")
				v := ""
				if len(kv) > 1 {
					v = kv[1]
				}

				rop.Params[kv[0]] = v
			}
		}
	}

	return rop
}

type Organization struct {
	Id string `json:"_id"`
}

type Organizations struct {
	Items []*Organization `json:"organizations"`
}

type Workspace struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Workspaces struct {
	Items []*Workspace `json:"workspaces"`
}

type EncryptedKey struct {
	Encryptedkey string `json:"encryptedkey"`
	Nonce        string `json:"nonce"`
	Receiver     string `json:"receiver"`
	Sender       Sender `json:"sender"`
	Workspace    string `json:"workspace"`
}

type Sender struct {
	PublicKey string `json:"publicKey"`
}

const DefaulSecretEnv = "prod"

const StatusOK = 0
