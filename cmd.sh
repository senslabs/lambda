fission env create --spec --name go --image dozeelabs/fission-go-env --builder dozeelabs/fission-go-builder:20200430 --mincpu 40 --maxcpu 80 --minmemory 64 --maxmemory 128 --poolsize 3

fission function create --spec --name request-otp --env go --src "sens/ws/*" --entrypoint RequestOtp
fission route create --spec --method POST --url /api/otp/request --function request-otp --name request-otp

fission function create --spec --name verify-otp --env go --src "sens/ws/*" --entrypoint VerifyOtp
fission route create --spec --method POST --url /api/otp/verify --function verify-otp --name verify-otp

fission function create --spec --name verify-auth --env go --src "sens/ws/*" --entrypoint VerifyAuth
fission route create --spec --method POST --url /api/auth/verify --function verify-auth --name verify-auth

fission function create --spec --name delete-auth --env go --src "sens/ws/*" --entrypoint DeleteAuth
fission route create --spec --method POST --url /api/auths/delete --function delete-auth --name delete-auth

fission function create --spec --name add-org --env go --src "sens/ws/*" --entrypoint AddOrg
fission route create --spec --method POST --url /api/orgs/add --function add-org --name add-org

fission function create --spec --name create-org --env go --src "sens/ws/*" --entrypoint CreateOrg
fission route create --spec --method POST --url /api/orgs/create --function create-org --name create-org

fission function create --spec --name create-op --env go --src "sens/ws/*" --entrypoint CreateOp
fission route create --spec --method POST --url /api/ops/create --function create-op --name create-op

fission function create --spec --name add-op --env go --src "sens/ws/*" --entrypoint AddOp
fission route create --spec --method POST --url /api/ops/add --function add-op --name add-op

fission function create --spec --name create-user --env go --src "sens/ws/*" --entrypoint CreateUser
fission route create --spec --method POST --url /api/users/create --function create-user --name create-user

fission function create --spec --name add-user --env go --src "sens/ws/*" --entrypoint AddUser
fission route create --spec --method POST --url /api/users/add --function add-user --name add-user

fission function create --spec --name refresh-token --env go --src "sens/ws/*" --entrypoint RefreshToken
fission route create --spec --method GET --url /api/tokens/refresh --function refresh-token --name refresh-token

fission function create --spec --name change-role-to-op --env go --src "sens/ws/*" --entrypoint ChangeRoleToOp
fission route create --spec --method POST --url /api/roles/update/op --function change-role-to-op --name change-role-to-op

fission function create --spec --name change-role-to-auth --env go --src "sens/ws/*" --entrypoint ChangeRoleToAuth
fission route create --spec --method POST --url /api/roles/update/auth --function change-role-to-auth --name change-role-to-auth

fission function create --spec --name list-orgs --env go --src "sens/ws/*" --entrypoint ListOrgs
fission route create --spec --method GET --url /api/orgs/list --function list-orgs --name list-orgs

fission function create --spec --name list-ops --env go --src "sens/ws/*" --entrypoint ListOps
fission route create --spec --method GET --url /api/ops/list --function list-ops --name list-ops

fission function create --spec --name list-users --env go --src "sens/ws/*" --entrypoint ListUsers
fission route create --spec --method GET --url /api/users/list --function list-users --name list-users

fission function create --spec --name list-devices --env go --src "sens/ws/*" --entrypoint ListDevices
fission route create --spec --method GET --url /api/devices/list --function list-devices --name list-devices

fission function create --spec --name list-users-summary --env go --src "sens/ws/*" --entrypoint ListUserSummary
fission route create --spec --method GET --url /api/users/summary --function list-users-summary --name list-users-summary

fission function create --spec --name list-users-sleeps --env go --src "sens/ws/*" --entrypoint ListUserSleeps
fission route create --spec --method GET --url /api/users/sleeps --function list-users-sleeps --name list-users-sleeps

fission function create --spec --name list-users-meditations --env go --src "sens/ws/*" --entrypoint ListUserMeditations
fission route create --spec --method GET --url /api/users/meditations --function list-users-meditations --name list-users-meditations

fission function create --spec --name list-activities --env go --src "sens/ws/*" --entrypoint ListActivities
fission route create --spec --method GET --url /api/activities/list --function list-activities --name list-activities

fission function create --spec --name get-sleeps-summary --env go --src "sens/ws/*" --entrypoint GetSleepsSummary
fission route create --spec --method GET --url /api/sleeps/summary/get --function get-sleeps-summary --name get-sleeps-summary

fission function create --spec --name get-meditations-summary --env go --src "sens/ws/*" --entrypoint GetMeditationsSummary
fission route create --spec --method GET --url /api/meditations/summary/get --function get-meditations-summary --name get-meditations-summary

fission function create --spec --name list-alerts --env go --src "sens/ws/*" --entrypoint ListAlerts
fission route create --spec --method GET --url /api/alerts/list --function list-alerts --name list-alerts

fission function create --spec --name list-latest-alerts --env go --src "sens/ws/*" --entrypoint ListLatestAlerts
fission route create --spec --method GET --url /api/alerts/latest/list --function list-latest-alerts --name list-latest-alerts

fission function create --spec --name create-api-key --env go --src "sens/ws/*" --entrypoint CreateKey
fission route create --spec --method POST --url /api/keys/create --function create-api-key --name create-api-key

fission function create --spec --name list-api-keys --env go --src "sens/ws/*" --entrypoint ListKeys
fission route create --spec --method POST --url /api/keys/list --function list-api-keys --name list-api-keys

fission function create --spec --name create-device --env go --src "sens/ws/*" --entrypoint CreateDevice
fission route create --spec --method POST --url /api/devices/create --function create-device --name create-device

fission function create --spec --name register-device --env go --src "sens/ws/*" --entrypoint RegisterDevice
fission route create --spec --method POST --url /api/devices/register --function register-device --name register-device

fission function create --spec --name unregister-device --env go --src "sens/ws/*" --entrypoint UnregisterDevice
fission route create --spec --method POST --url /api/devices/unregister --function unregister-device --name unregister-device

fission function create --spec --name pair-device --env go --src "sens/ws/*" --entrypoint PairDevice
fission route create --spec --method POST --url /api/devices/pair --function pair-device --name pair-device

fission function create --spec --name unpair-device --env go --src "sens/ws/*" --entrypoint UnpairDevice
fission route create --spec --method POST --url /api/devices/unpair --function unpair-device --name unpair-device

fission function create --spec --name get-session --env go --src "sens/ws/*" --entrypoint GetSession
fission route create --spec --method POST --url /api/sessions/{sessionId}/get --function get-session --name get-session

fission function create --spec --name get-quarter-activity-counts --env go --src "sens/ws/*" --entrypoint GetQuarterActivityCounts
fission route create --spec --method GET --url /api/org/quarter-activities/counts --function get-quarter-activity-counts --name get-quarter-activity-counts

fission function create --spec --name get-user-session --env go --src "sens/ws/*" --entrypoint GetUserSession
fission route create --spec --method GET --url /api/sessions/{id}/get --function get-user-session --name get-user-session

fission function create --spec --name get-user-session-events --env go --src "sens/ws/*" --entrypoint GetUserSessionEvents
fission route create --spec --method GET --url /api/sessions/{id}/events/get --function get-user-session-events --name get-user-session-events

fission function create --spec --name add-device-properties --env go --src "sens/ws/*" --entrypoint AddDeviceProperties
fission route create --spec --method POST --url /api/devices/properties/add --function add-device-properties --name add-device-properties

fission function create --spec --name get-device-properties --env go --src "sens/ws/*" --entrypoint GetDeviceProperties
fission route create --spec --method GET --url /api/devices/{id}/properties/get --function get-device-properties --name get-device-properties

fission function create --spec --name create-alert-rule --env go --src "sens/ws/*" --entrypoint CreateAlertRule
fission route create --spec --method POST --url /api/alert-rules/create --function create-alert-rule --name create-alert-rule

fission function create --spec --name create-alert-escalation --env go --src "sens/ws/*" --entrypoint CreateAlertEscalation
fission route create --spec --method POST --url /api/alert-escalations/create --function create-alert-escalation --name create-alert-escalation

fission function create --spec --name update-alert-rule --env go --src "sens/ws/*" --entrypoint UpdateAlertRule
fission route create --spec --method POST --url /api/alert-rules/{id}/update --function update-alert-rule --name update-alert-rule

fission function create --spec --name update-alert-escalation --env go --src "sens/ws/*" --entrypoint UpdateAlertEscalation
fission route create --spec --method POST --url /api/alert-escalations/{id}/update --function update-alert-escalation --name update-alert-escalation

fission function create --spec --name update-alert --env go --src "sens/ws/*" --entrypoint UpdateAlert
fission route create --spec --method POST --url /api/alerts/{id}/update --function update-alert --name update-alert

fission function create --spec --name update-org-properties --env go --src "sens/ws/*" --entrypoint UpdateOrgProperties
fission route create --spec --method POST --url /api/orgs/properties/update --function update-org-properties --name update-org-properties

fission function create --spec --name update-op-properties --env go --src "sens/ws/*" --entrypoint UpdateOpProperties
fission route create --spec --method POST --url /api/ops/properties/update --function update-op-properties --name update-op-properties

fission function create --spec --name update-user-properties --env go --src "sens/ws/*" --entrypoint UpdateUserProperties
fission route create --spec --method POST --url /api/users/{id}/properties/update --function update-user-properties --name update-user-properties

fission function create --spec --name list-org-properties --env go --src "sens/ws/*" --entrypoint ListOrgProperties
fission route create --spec --method GET --url /api/orgs/{id}/properties/list --function list-org-properties --name list-org-properties

fission function create --spec --name list-op-properties --env go --src "sens/ws/*" --entrypoint ListOpProperties
fission route create --spec --method GET --url /api/ops/{id}/properties/list --function list-op-properties --name list-op-properties

fission function create --spec --name list-user-properties --env go --src "sens/ws/*" --entrypoint ListUserProperties
fission route create --spec --method GET --url /api/users/{id}/properties/list --function list-user-properties --name list-user-properties

fission function create --spec --name list-alert-rules --env go --src "sens/ws/*" --entrypoint ListAlertRules
fission route create --spec --method GET --url /api/alert-rules/list --function list-alert-rules --name list-alert-rules

fission function create --spec --name list-alert-escalations --env go --src "sens/ws/*" --entrypoint ListAlertEscalations
fission route create --spec --method GET --url /api/alert-escalations/list --function list-alert-escalations --name list-alert-escalations

fission function create --spec --name get-users-sleep-count --env go --src "sens/ws/*" --entrypoint GetUsersSleepCount
fission route create --spec --method GET --url /api/users/sleeps/count --function get-users-sleep-count --name get-users-sleep-count

fission function create --spec --name get-session-durations --env go --src "sens/ws/*" --entrypoint GetSessionDurations
fission route create --spec --method GET --url /api/sessions/durations/get --function get-session-durations --name get-session-durations

fission function create --spec --name get-user-reports --env go --src "sens/ws/*" --entrypoint GetUserReports
fission route create --spec --method GET --url /api/users/reports/get --function get-user-reports --name get-user-reports

fission function create --spec --name request-user-reports --env go --src "sens/ws/*" --entrypoint RequestUserReports
fission route create --spec --method POST --url /api/users/{id}/report/request --function request-user-reports --name request-user-reports

fission function create --spec --name get-sleep-trend --env go --src "sens/ws/*" --entrypoint GetSleepTrend
fission route create --spec --method GET --url /api/users/{id}/sleep-trend/get --function get-sleep-trend --name get-sleep-trend

fission function create --spec --name get-user-dated-session --env go --src "sens/ws/*" --entrypoint GetUserDatedSession
fission route create --spec --method GET --url /api/users/{id}/dated-sessions/get --function get-user-dated-session --name get-user-dated-session

fission function create --spec --name get-longest-sleep-trend --env go --src "sens/ws/*" --entrypoint GetLongestSleepTrend
fission route create --spec --method GET --url /api/users/{id}/longest-trend/get --function get-longest-sleep-trend --name get-longest-sleep-trend

fission function create --spec --name list-user-baselines --env go --src "sens/ws/*" --entrypoint ListUserBaselines
fission route create --spec --method GET --url /api/users/{id}/baselines/list --function list-user-baselines --name list-user-baselines

