#!/bin/bash

convert_to_memory_unit() {
    local num=$1
    local unit="B"
    if [[ $num -ge 1073741824 ]]; then
        num=`echo $num | awk '{ printf("%.2lf",$1/1073741824) }'`
        unit="GB"
    elif [[ $num -ge 1048576 ]]; then
        num=`echo $num | awk '{ printf("%.2lf",$1/1048576) }'`
        unit="MB"
    elif [[ $num -ge 1024 ]]; then
	num=`echo $num | awk '{ printf("%.2lf",$1/1024) }'`
        unit="KB"
    fi
    echo "$num$unit"
}


>image.size.tmp.txt
>image.size.raw.txt
>image.size.txt

for path in `ls images/*.tar.gz`
do
        image=`basename $path`
        rm -rf tmp
        mkdir tmp
        cp  images/$image tmp/$image

        cd tmp
        size=`ls -l | awk '{ print $5 }' | tail --line 1`
        echo $image
        tar -xzf $image
        name=`cat manifest.json  | awk -F"RepoTags" '{ print $2 }' | awk -F"\"" '{ print $3 }'`
        cd ..

        echo -e $size"\t"$name >> image.size.tmp.txt
done

sort -k1 -nr image.size.tmp.txt > image.size.raw.txt
rm image.size.tmp.txt

while read size image
do
	echo -e $(convert_to_memory_unit $size)"\t"$image >> image.size.txt
done < image.size.raw.txt
