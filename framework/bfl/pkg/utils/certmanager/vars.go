package certmanager

var (
	ReDownloadCertificateAPIFormat = "http://bfl.%s/bfl/backend/v1/re-download-cert"

	DefaultAheadRenewalCertDays = -7

	CertExpiredDateTimeLayout = "2006-01-02T15:04:05Z"

	ReDownloadCertCronJobName = "expired-download-cert"

	ReDownloadCertCronJobScheduleFormat = "%d %d %d %d ?"
)
