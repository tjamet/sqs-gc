package main

import (
	"flag"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := sqs.New(sess)
	cw := cloudwatch.New(sess)
	maxAge := flag.Float64("max-age", 1000000, "The threshold (seconds) after which if a queue with a message this old will be listed for deletion")
	real := flag.Bool("delete-really", false, "Provide this flag to actually perform the deletion")
	pattern := flag.String("queues", "*", "Provide a pattern queues to delete must match. Matching will be done using the fnmatch pattern (see https://golang.org/pkg/path/filepath/#Match)")
	flag.Parse()

	queues, err := svc.ListQueues(&sqs.ListQueuesInput{})
	if err != nil {
		fmt.Println(err)
	} else {
		for _, queueURL := range queues.QueueUrls {
			u, err := url.Parse(*queueURL)
			if err != nil {
				fmt.Println(err)
			} else {
				queueName := path.Base(u.Path)
				m, err := filepath.Match(*pattern, queueName)
				if err != nil {
					fmt.Printf("can't evaluate queue %s: %s\n", queueName, err.Error())
				} else if m {
					metrics, err := cw.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
						Statistics: []*string{aws.String("Maximum")},
						Namespace:  aws.String("AWS/SQS"),
						Dimensions: []*cloudwatch.Dimension{
							{
								Name:  aws.String("QueueName"),
								Value: aws.String(queueName),
							},
						},
						MetricName: aws.String("ApproximateAgeOfOldestMessage"),
						StartTime:  aws.Time(time.Now().Add(-10 * time.Minute)),
						EndTime:    aws.Time(time.Now()),
						Period:     aws.Int64(10 * 60), // 10 minutes to have 1 datapoints
					})
					if err != nil {
						fmt.Println(err)
					} else {
						if len(metrics.Datapoints) > 0 {
							dp := metrics.Datapoints[0]
							if *dp.Unit != "Seconds" {
								fmt.Printf("skipping queue %s, the reported statistics are in %s instead of the expected Seconds\n", queueName, *dp.Unit)
								continue
							}
							if *dp.Maximum > *maxAge {
								if *real {
									response, err := svc.DeleteQueue(&sqs.DeleteQueueInput{
										QueueUrl: queueURL,
									})
									if err != nil {
										fmt.Println(err)
									} else {
										fmt.Printf(
											"queue %s had an oldest message age of %.0f %s more than the %.0f accepted, it has been deleted: %s\n",
											queueName, *dp.Maximum, *dp.Unit, *maxAge, response.GoString())
									}
								} else {
									fmt.Printf("would delete queue %s with an oldest message age of %.0f %s\n", queueName, *dp.Maximum, *dp.Unit)
								}
							} else {
								fmt.Printf("queue %s has an oldest message age of %.0f %s, keeping it\n", queueName, *dp.Maximum, *dp.Unit)
							}

						} else {
							fmt.Printf("no oldest message age statistic for queue %s, skipping it\n", queueName)
						}
					}
				} else {
					fmt.Printf("queue %s does not match pattern %s, skipping it\n", queueName, *pattern)
				}
			}
		}
	}
}
