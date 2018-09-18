# sqs-gc
A simple garbage collector for SQS queues

List all SQS queues in your account, fetches the cloudwatch metrics and deletes the ones for which
the ApproximateAgeOfOldestMessage is greater than the provided value.

The program would run once only.

# Usage

Download it: `go get github.com/tjamet/sqs-gc`

run it:

```bash
sqs-gc --help
Usage of sqs-gs:
  -delete-really
        Provide this flag to actually perform the deletion
  -max-age float
        The threshold (seconds) after which if a queue with a message this old will be listed for deletion (default 1e+06)
  -queues string
        Provide a pattern queues to delete must match. Matching will be done using the fnmatch pattern (see https://golang.org/pkg/path/filepath/#Match) (default "*")
```

# Run periodically

Run the script in a cron tab, or a kubernetes cronjob to periodically cleanup your SQS queues