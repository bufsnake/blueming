#!/bin/bash  
  
for i in $(find ./output)  
do   
    if [[ $(file $i | grep "HTML") == "" &&  $(file $i | grep "empty") == "" && $(file $i | grep "JSON") == "" ]]
    then
        echo $i
    else
        rm -rf $i
    fi
done 
