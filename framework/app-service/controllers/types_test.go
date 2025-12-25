package controllers

import (
	"encoding/json"

	appv1alpha1 "github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("mergeEntrances", func() {
	It("should return incoming when existing is empty", func() {
		incoming := []appv1alpha1.Entrance{
			{Name: "web", Host: "web-svc", Port: 8080, AuthLevel: "public"},
		}

		result := mergeEntrances(nil, incoming)

		Expect(result).To(Equal(incoming))
	})

	It("should return incoming when existing is empty slice", func() {
		existing := []appv1alpha1.Entrance{}
		incoming := []appv1alpha1.Entrance{
			{Name: "web", Host: "web-svc", Port: 8080, AuthLevel: "public"},
		}

		result := mergeEntrances(existing, incoming)

		Expect(result).To(Equal(incoming))
	})

	It("should preserve authLevel from existing entrances", func() {
		existing := []appv1alpha1.Entrance{
			{Name: "web", Host: "old-svc", Port: 80, AuthLevel: "private"},
		}
		incoming := []appv1alpha1.Entrance{
			{Name: "web", Host: "new-svc", Port: 8080, AuthLevel: "public"},
		}

		result := mergeEntrances(existing, incoming)

		Expect(result).To(HaveLen(1))
		Expect(result[0].Name).To(Equal("web"))
		Expect(result[0].Host).To(Equal("new-svc"))
		Expect(result[0].Port).To(Equal(int32(8080)))
		Expect(result[0].AuthLevel).To(Equal("private")) // preserved from existing
	})

	It("should update other fields from incoming", func() {
		existing := []appv1alpha1.Entrance{
			{Name: "web", Host: "old-svc", Port: 80, AuthLevel: "private", Title: "Old Title", Icon: "old-icon"},
		}
		incoming := []appv1alpha1.Entrance{
			{Name: "web", Host: "new-svc", Port: 8080, AuthLevel: "public", Title: "New Title", Icon: "new-icon"},
		}

		result := mergeEntrances(existing, incoming)

		Expect(result).To(HaveLen(1))
		Expect(result[0].Host).To(Equal("new-svc"))
		Expect(result[0].Port).To(Equal(int32(8080)))
		Expect(result[0].Title).To(Equal("New Title"))
		Expect(result[0].Icon).To(Equal("new-icon"))
		Expect(result[0].AuthLevel).To(Equal("private")) // preserved
	})

	It("should handle new entrance not in existing", func() {
		existing := []appv1alpha1.Entrance{
			{Name: "web", Host: "web-svc", Port: 80, AuthLevel: "private"},
		}
		incoming := []appv1alpha1.Entrance{
			{Name: "web", Host: "web-svc", Port: 80, AuthLevel: "public"},
			{Name: "api", Host: "api-svc", Port: 3000, AuthLevel: "public"},
		}

		result := mergeEntrances(existing, incoming)

		Expect(result).To(HaveLen(2))
		// web entrance preserves authLevel
		Expect(result[0].Name).To(Equal("web"))
		Expect(result[0].AuthLevel).To(Equal("private"))
		// api entrance uses incoming authLevel (no existing)
		Expect(result[1].Name).To(Equal("api"))
		Expect(result[1].AuthLevel).To(Equal("public"))
	})

	It("should handle removed entrance from existing", func() {
		existing := []appv1alpha1.Entrance{
			{Name: "web", Host: "web-svc", Port: 80, AuthLevel: "private"},
			{Name: "api", Host: "api-svc", Port: 3000, AuthLevel: "private"},
		}
		incoming := []appv1alpha1.Entrance{
			{Name: "web", Host: "web-svc", Port: 80, AuthLevel: "public"},
		}

		result := mergeEntrances(existing, incoming)

		Expect(result).To(HaveLen(1))
		Expect(result[0].Name).To(Equal("web"))
		Expect(result[0].AuthLevel).To(Equal("private"))
	})

	It("should handle multiple entrances correctly", func() {
		existing := []appv1alpha1.Entrance{
			{Name: "web", Host: "web-svc", Port: 80, AuthLevel: "private"},
			{Name: "admin", Host: "admin-svc", Port: 9000, AuthLevel: "internal"},
		}
		incoming := []appv1alpha1.Entrance{
			{Name: "web", Host: "web-new", Port: 8080, AuthLevel: "public"},
			{Name: "admin", Host: "admin-new", Port: 9090, AuthLevel: "public"},
			{Name: "api", Host: "api-svc", Port: 3000, AuthLevel: "public"},
		}

		result := mergeEntrances(existing, incoming)

		Expect(result).To(HaveLen(3))
		// web: authLevel preserved
		Expect(result[0].Name).To(Equal("web"))
		Expect(result[0].Host).To(Equal("web-new"))
		Expect(result[0].AuthLevel).To(Equal("private"))
		// admin: authLevel preserved
		Expect(result[1].Name).To(Equal("admin"))
		Expect(result[1].Host).To(Equal("admin-new"))
		Expect(result[1].AuthLevel).To(Equal("internal"))
		// api: new entrance, uses incoming authLevel
		Expect(result[2].Name).To(Equal("api"))
		Expect(result[2].AuthLevel).To(Equal("public"))
	})
})

var _ = Describe("mergePolicySettings", func() {
	It("should return existing when incoming is empty", func() {
		existing := `{"calibreweb-svc":{"default_policy":"system","sub_policies":null,"one_time":false,"valid_duration":0}}`

		result := mergePolicySettings(existing, "")

		Expect(result).To(Equal(existing))
	})

	It("should return incoming when existing is empty", func() {
		incoming := `{"calibreweb-svc":{"default_policy":"system","one_time":false,"sub_policies":[{"one_time":true,"policy":"one_factor","uri":"/api/send","valid_duration":0}],"valid_duration":0}}`

		result := mergePolicySettings("", incoming)

		Expect(result).To(Equal(incoming))
	})

	It("should preserve default_policy from existing", func() {
		existing := `{"calibreweb-svc":{"default_policy":"private","sub_policies":null,"one_time":false,"valid_duration":0}}`
		incoming := `{"calibreweb-svc":{"default_policy":"system","sub_policies":null,"one_time":false,"valid_duration":0}}`

		result := mergePolicySettings(existing, incoming)

		var resultPolicy map[string]applicationSettingsPolicy
		err := json.Unmarshal([]byte(result), &resultPolicy)
		Expect(err).NotTo(HaveOccurred())

		Expect(resultPolicy["calibreweb-svc"].DefaultPolicy).To(Equal("private"))
	})

	It("should preserve sub_policies from existing", func() {
		existing := `{"calibreweb-svc":{"default_policy":"system","sub_policies":[{"uri":"/api/send","policy":"one_factor","one_time":true,"valid_duration":0}],"one_time":false,"valid_duration":0}}`
		incoming := `{"calibreweb-svc":{"default_policy":"system","sub_policies":[{"uri":"/api/new","policy":"two_factor","one_time":false,"valid_duration":3600}],"one_time":false,"valid_duration":0}}`

		result := mergePolicySettings(existing, incoming)

		var resultPolicy map[string]applicationSettingsPolicy
		err := json.Unmarshal([]byte(result), &resultPolicy)
		Expect(err).NotTo(HaveOccurred())

		// sub_policies preserved from existing
		Expect(resultPolicy["calibreweb-svc"].SubPolicies).To(HaveLen(1))
		Expect(resultPolicy["calibreweb-svc"].SubPolicies[0].URI).To(Equal("/api/send"))
		Expect(resultPolicy["calibreweb-svc"].SubPolicies[0].Policy).To(Equal("one_factor"))
		Expect(resultPolicy["calibreweb-svc"].SubPolicies[0].OneTime).To(BeTrue())
	})

	It("should preserve both default_policy and sub_policies from existing", func() {
		existing := `{"calibreweb-svc":{"default_policy":"public","sub_policies":[{"uri":"/api/send","policy":"one_factor","one_time":true,"valid_duration":0}],"one_time":false,"valid_duration":0}}`
		incoming := `{"calibreweb-svc":{"default_policy":"system","sub_policies":[{"uri":"/api/new","policy":"two_factor","one_time":false,"valid_duration":3600}],"one_time":true,"valid_duration":1800}}`

		result := mergePolicySettings(existing, incoming)

		var resultPolicy map[string]applicationSettingsPolicy
		err := json.Unmarshal([]byte(result), &resultPolicy)
		Expect(err).NotTo(HaveOccurred())

		// default_policy preserved from existing
		Expect(resultPolicy["calibreweb-svc"].DefaultPolicy).To(Equal("public"))
		// sub_policies preserved from existing
		Expect(resultPolicy["calibreweb-svc"].SubPolicies).To(HaveLen(1))
		Expect(resultPolicy["calibreweb-svc"].SubPolicies[0].URI).To(Equal("/api/send"))
		// other fields use incoming
		Expect(resultPolicy["calibreweb-svc"].OneTime).To(BeTrue())
		Expect(resultPolicy["calibreweb-svc"].Duration).To(Equal(int32(1800)))
	})

	It("should preserve null sub_policies from existing", func() {
		existing := `{"calibreweb-svc":{"default_policy":"system","sub_policies":null,"one_time":false,"valid_duration":0}}`
		incoming := `{"calibreweb-svc":{"default_policy":"system","sub_policies":[{"uri":"/api/send","policy":"one_factor","one_time":true,"valid_duration":0}],"one_time":false,"valid_duration":0}}`

		result := mergePolicySettings(existing, incoming)

		var resultPolicy map[string]applicationSettingsPolicy
		err := json.Unmarshal([]byte(result), &resultPolicy)
		Expect(err).NotTo(HaveOccurred())

		// sub_policies preserved as null from existing
		Expect(resultPolicy["calibreweb-svc"].SubPolicies).To(BeNil())
	})

	It("should add new entrance policy with incoming values", func() {
		existing := `{"calibreweb-svc":{"default_policy":"private","sub_policies":[{"uri":"/old","policy":"one_factor","one_time":false,"valid_duration":0}],"one_time":false,"valid_duration":0}}`
		incoming := `{"calibreweb-svc":{"default_policy":"system","sub_policies":null,"one_time":false,"valid_duration":0},"api-svc":{"default_policy":"public","sub_policies":[{"uri":"/api","policy":"two_factor","one_time":true,"valid_duration":3600}],"one_time":false,"valid_duration":0}}`

		result := mergePolicySettings(existing, incoming)

		var resultPolicy map[string]applicationSettingsPolicy
		err := json.Unmarshal([]byte(result), &resultPolicy)
		Expect(err).NotTo(HaveOccurred())

		// existing entry: preserved default_policy and sub_policies
		Expect(resultPolicy["calibreweb-svc"].DefaultPolicy).To(Equal("private"))
		Expect(resultPolicy["calibreweb-svc"].SubPolicies).To(HaveLen(1))
		Expect(resultPolicy["calibreweb-svc"].SubPolicies[0].URI).To(Equal("/old"))
		// new entry: uses incoming values
		Expect(resultPolicy).To(HaveKey("api-svc"))
		Expect(resultPolicy["api-svc"].DefaultPolicy).To(Equal("public"))
		Expect(resultPolicy["api-svc"].SubPolicies).To(HaveLen(1))
		Expect(resultPolicy["api-svc"].SubPolicies[0].URI).To(Equal("/api"))
	})

	It("should delete entrance policy not in incoming", func() {
		existing := `{"calibreweb-svc":{"default_policy":"system","sub_policies":null,"one_time":false,"valid_duration":0},"api-svc":{"default_policy":"public","sub_policies":null,"one_time":false,"valid_duration":0}}`
		incoming := `{"calibreweb-svc":{"default_policy":"system","sub_policies":null,"one_time":false,"valid_duration":0}}`

		result := mergePolicySettings(existing, incoming)

		var resultPolicy map[string]applicationSettingsPolicy
		err := json.Unmarshal([]byte(result), &resultPolicy)
		Expect(err).NotTo(HaveOccurred())

		Expect(resultPolicy).To(HaveKey("calibreweb-svc"))
		Expect(resultPolicy).NotTo(HaveKey("api-svc"))
	})

	It("should handle add, preserve, and delete together", func() {
		existing := `{
			"web-svc":{"default_policy":"private","sub_policies":[{"uri":"/web","policy":"one_factor","one_time":false,"valid_duration":0}],"one_time":false,"valid_duration":0},
			"admin-svc":{"default_policy":"internal","sub_policies":[{"uri":"/admin","policy":"two_factor","one_time":true,"valid_duration":3600}],"one_time":false,"valid_duration":0},
			"legacy-svc":{"default_policy":"public","sub_policies":null,"one_time":false,"valid_duration":0}
		}`
		incoming := `{
			"web-svc":{"default_policy":"system","sub_policies":null,"one_time":true,"valid_duration":1800},
			"admin-svc":{"default_policy":"system","sub_policies":[{"uri":"/new","policy":"one_factor","one_time":false,"valid_duration":0}],"one_time":false,"valid_duration":0},
			"api-svc":{"default_policy":"system","sub_policies":[{"uri":"/api","policy":"one_factor","one_time":false,"valid_duration":0}],"one_time":false,"valid_duration":0}
		}`

		result := mergePolicySettings(existing, incoming)

		var resultPolicy map[string]applicationSettingsPolicy
		err := json.Unmarshal([]byte(result), &resultPolicy)
		Expect(err).NotTo(HaveOccurred())

		// web-svc: default_policy and sub_policies preserved, other fields from incoming
		Expect(resultPolicy["web-svc"].DefaultPolicy).To(Equal("private"))
		Expect(resultPolicy["web-svc"].SubPolicies).To(HaveLen(1))
		Expect(resultPolicy["web-svc"].SubPolicies[0].URI).To(Equal("/web"))
		Expect(resultPolicy["web-svc"].OneTime).To(BeTrue())
		Expect(resultPolicy["web-svc"].Duration).To(Equal(int32(1800)))
		// admin-svc: default_policy and sub_policies preserved
		Expect(resultPolicy["admin-svc"].DefaultPolicy).To(Equal("internal"))
		Expect(resultPolicy["admin-svc"].SubPolicies).To(HaveLen(1))
		Expect(resultPolicy["admin-svc"].SubPolicies[0].URI).To(Equal("/admin"))
		Expect(resultPolicy["admin-svc"].SubPolicies[0].Policy).To(Equal("two_factor"))
		// api-svc: new entry, uses incoming values
		Expect(resultPolicy).To(HaveKey("api-svc"))
		Expect(resultPolicy["api-svc"].DefaultPolicy).To(Equal("system"))
		Expect(resultPolicy["api-svc"].SubPolicies).To(HaveLen(1))
		Expect(resultPolicy["api-svc"].SubPolicies[0].URI).To(Equal("/api"))
		// legacy-svc: deleted (not in incoming)
		Expect(resultPolicy).NotTo(HaveKey("legacy-svc"))
	})

	It("should return incoming when existing JSON is invalid", func() {
		existing := `invalid json`
		incoming := `{"calibreweb-svc":{"default_policy":"system","sub_policies":null,"one_time":false,"valid_duration":0}}`

		result := mergePolicySettings(existing, incoming)

		Expect(result).To(Equal(incoming))
	})

	It("should return existing when incoming JSON is invalid", func() {
		existing := `{"calibreweb-svc":{"default_policy":"system","sub_policies":null,"one_time":false,"valid_duration":0}}`
		incoming := `invalid json`

		result := mergePolicySettings(existing, incoming)

		Expect(result).To(Equal(existing))
	})
})
