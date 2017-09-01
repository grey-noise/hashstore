# Hash store

[![Build Status](https://travis-ci.org/grey-noise/hashstore.svg?branch=master)](https://travis-ci.org/grey-noise/hashstore)
![Build size](https://reposs.herokuapp.com/?path=grey-noise/hashstore)
[![Go Report Card](https://goreportcard.com/badge/github.com/grey-noise/hashstore)](https://goreportcard.com/report/github.com/grey-noise/hashstore)
[![Coverage Status](https://coveralls.io/repos/github/grey-noise/hashstore/badge.svg)](https://coveralls.io/github/grey-noise/hashstore)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A simple Go application for compare one structure of files (base of MD5 hash).

NAME:
   goHashStore-windows-amd64.exe - Hashstore application

USAGE:
   goHashStore-windows-amd64.exe [global options] command [command options] [arg
uments...]

VERSION:
   v1.8

COMMANDS:
     start, s    start location
     display, h  display runid
     error, r    error runid
     delete, h   delete runid
     list, a     list
     compare, a  compare runid1 runid2
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --db value               the database location (must have writing access) (de
fault: "hashstore.db")
   --pr value               the number of workers (default: 3)
   --help, -h               show help (default: false)
   --init-completion value  generate completion code. Value must be 'bash' or 'z
sh'
   --version, -v            print the version (default: false)
