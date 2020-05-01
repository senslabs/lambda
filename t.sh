fission env delete --name go
#fission env create --name go --image dozeelabs/fission-go-env --builder dozeelabs/fission-go-builder --mincpu 40 --maxcpu 80 --minmemory 64 --maxmemory 128 --poolsize 3
fission env create --name go --image fission/go-env-1.13 --builder fission/go-builder-1.13 --mincpu 40 --maxcpu 80 --minmemory 64 --maxmemory 128 --poolsize 3

#fission fn delete --name list-op-properties
#fission function create --name list-op-properties --env go --src "sens/ws/*" --entrypoint ListOpProperties

