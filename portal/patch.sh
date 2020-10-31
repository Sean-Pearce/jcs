#!/usr/bin/env bash

for patch in $(ls patches/)
do
	patch -p1 <patches/$patch
done
