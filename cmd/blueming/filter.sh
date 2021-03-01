#!/bin/bash  
  
for i in $(find ./output)  
do   
    if [[ $(file $i | grep $i": data") == "" && $(file $i | grep "image data") == "" && $(file $i | grep "HTML") == "" && $(file $i | grep "empty") == "" && $(file $i | grep "JSON") == "" && $(file $i | grep "text") == "" ]]
    then
        echo $i
        file $i
    else
        rm -rf $i
    fi
done 
