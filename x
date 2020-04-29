fission function create --spec --name list-alerts --env go --src "sens/ws/*" --entrypoint ListAlerts
fission route create --spec --method GET --url /api/alerts/list --function list-alerts --name list-alerts

fission function create --spec --name list-latest-alerts --env go --src "sens/ws/*" --entrypoint ListLatestAlerts
fission route create --spec --method GET --url /api/alerts/latest/list --function list-latest-alerts --name list-latest-alerts
