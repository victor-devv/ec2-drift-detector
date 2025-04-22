package detector

import (
	"github.com/sirupsen/logrus"
	"github.com/victor-devv/ec2-drift-detector/internal/aws"
	"github.com/victor-devv/ec2-drift-detector/internal/terraform"
	"github.com/victor-devv/ec2-drift-detector/pkg/utils"
)

// EC2Detector handles drift detection for EC2 instances
type EC2Detector struct {
	BaseDetector
	ec2Client  aws.EC2ClientImpl
	tfParser   terraform.Parser
	log        *logrus.Logger
	comparator *utils.Comparator
}

// NewEC2Detector creates a new EC2 detector
func NewEC2Detector(ec2Client *aws.EC2ClientImpl, tfParser terraform.Parser, log *logrus.Logger) *EC2Detector {
	return &EC2Detector{
		ec2Client:  *ec2Client,
		tfParser:   tfParser,
		log:        log,
		comparator: utils.NewComparator(),
	}
}
