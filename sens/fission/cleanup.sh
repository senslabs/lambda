for f in `fission fn list | awk '{print $1}'`
do
  if [ "$f" == "NAME" ]; then
    continue
  fi
  fission fn delete --name $f --verbosity=2
done

for f in `fission pkg list | awk '{print $1}'`
do
  if [ "$f" == "NAME" ]; then
    continue
  fi
  fission pkg delete --name $f --verbosity=2
done

for f in `fission route list | awk '{print $1}'`
do
  if [ "$f" == "NAME" ]; then
    continue
  fi
  fission pkg delete --name $f --verbosity=2
done