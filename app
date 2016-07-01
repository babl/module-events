#!/bin/sh

if [ ! -t 0 ]; then
  read name
  if [ -z $name ]; then
    unset name
  fi
fi

echo Hello ${name-World}
