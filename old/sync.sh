#!/usr/bin/env bash
 
shopt -s expand_aliases

fsDest='tuan@113.20.119.39:~/dev/'

build_exclude() {
  y=""
  while IFS= read -r line; do
    [[ $line =~ ^#.* ]] && continue
    [[ $line =~ ^\\.* ]] && continue
    [[ $line =~ ^$ ]] && continue
    [[ $line =~ ^/priv/google/ ]] && continue
    if [[ $line =~ ^/.* ]]; then
      line=$(echo $line | cut -c2-)
      [[ $line != */ ]] && line="$line"/
    fi
    y="$y --exclude='$line' "
  done < "../$1/.gitignore"
  echo $y
  # return x
}

sync() {
  options="
    --progress \
    --partial \
    --archive \
    --verbose \
    --compress \
    --delete \
    --keep-dirlinks \
    --rsh=/usr/bin/ssh \
    --exclude='.git/' \
    $(build_exclude $1)"
  echo $options
  eval rsync $options ../$1 $fsDest
    # --exclude-from '../pancake_v2/.gitignore' \
}

watch() {
  sync $1; fswatch \
    --print0 \
    --one-per-batch \
    --recursive \
    --exclude=.elixir_ls \
    "../$1" | while read -d "" event; do \ 
      sync $1 \
    ; done
}
# watch "pancake_v2"
while IFS= read -r line; do
  eval watch $line &
done < "project_name"
wait
