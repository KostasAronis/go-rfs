#TODO: help() and usage example

echo "Graphing files in $1"
timestamp=$(date +"%d-%m-%Y_%H-%M-%S")
GoVector -log_dir "$1" -log_type Shiviz -outfile "$1/shiz.log"
go run /c/Code/DistributedSystemsAssignment/go-rfs/cmd/graph/graph.go "$1"
dot -Tjpeg -O `ls ./dockerenv/logs/*.dot`
echo "storing in out/$timestamp"
mkdir "out"
mv "$1" "out/$timestamp"