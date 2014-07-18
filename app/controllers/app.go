package controllers

import "strings"
import "sort"
import "github.com/revel/revel"
import "github.com/mitchellh/goamz/aws"
import "github.com/mitchellh/goamz/ec2"

type InstanceData struct {
    DisplayName string
    DisplayMetadata string
    InternalId string
    PhysicalEnvironment string
    PhysicalSubEnvironment string
    LogicalEnvironment string
    PrivateAddress string
    PublicAddress string
    Tags map[string]string
    Metadata map[string]string
}

func NewInstanceFromEc2(ei ec2.Instance) *InstanceData {
    i := &instance{}

    displayName := getTag(ei, "fqdn")
    if displayName == "" {
        displayName = getTag(ei, "Name")
        if displayName == "" {
            displayName = ei.InstanceId
        }
    }

    i.DisplayName = displayName
    i.InternalId = ei.InstanceId
    i.PrivateAddress = ei.PrivateIpAddress
    i.PublicAddress = ei.PublicIpAddress

    i.PhysicalEnvironment = "AWS"
    i.PhysicalSubEnvironment = ei.AvailZone

    logicalEnv := getTag(ei, "environment")
    if logicalEnv == "" {
        logicalEnv = "N/A"
    }

    i.LogicalEnvironment = logicalEnv

    i.Tags = getTagsAsDictionary(ei)

    i.Metadata = make(map[string]string)
    i.Metadata["VpcId"] = ei.VpcId
    i.Metadata["SecurityGroups"] = getSecurityGroups(ei)

    return i
}

type instanceCollection []*InstanceData

func (ic instanceCollection) Len() int {
    return len(ic)
}

func (ic instanceCollection) Swap(i, j int) {
    ic[i], ic[j] = ic[j], ic[i]
}

func (ic instanceCollection) Less(i, j int) bool {
    return ic[i].DisplayName < ic[j].DisplayName
}

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func (c App) Instances() revel.Result {
    awsKeyId, found := revel.Config.String("aws.key_id")
    if !found  {
        panic("no key id")
    }

    awsKey, found := revel.Config.String("aws.key")
    if !found {
        panic("no key")
    }

    auth, err := aws.GetAuth(awsKeyId, awsKey)
    if err != nil {
        panic(err)
    }

    compute := ec2.New(auth, aws.USEast)

    rawInstances, err := compute.Instances(nil, nil)
    if err != nil {
        panic(err)
    }

    instances := make(instanceCollection, 0)

    for _, r := range rawInstances.Reservations {
        for _, i := range r.Instances {
            // skip terminated instances
            if i.State.Code == 48 {
                continue
            }
            ni := NewInstanceFromEc2(i)
            instances = append(instances, ni)
        }
    }

    sort.Sort(instances)

    c.RenderArgs["instances"] = instances

    return c.Render()
}

func getTag(instance ec2.Instance, tagName string) string {
    for _, tag := range instance.Tags {
        if tag.Key == tagName {
            return tag.Value
        }
    }

    return ""
}

func getTagsAsDictionary(instance ec2.Instance) map[string]string {
    m := make(map[string]string)

    for _, tag := range instance.Tags {
        m[tag.Key] = tag.Value
    }

    return m
}

func getSecurityGroups(instance ec2.Instance) string {
    sgNames := make([]string, 0)

    for _, sg := range instance.SecurityGroups {
        sgNames = append(sgNames, sg.Name)
    }

    return strings.Join(sgNames, ", ")
}
