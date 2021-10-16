package main

import (
	"bytes"
	"errors"
	"fmt"
	//	apb "golang.conradwood.net/apis/autodeployer"
	pb "golang.conradwood.net/apis/deploymonkey"
)

type AppDiff struct {
	Was *pb.ApplicationDefinition
	Is  *pb.ApplicationDefinition
}
type Diff struct {
	AppDiffs []AppDiff
}

// describe in human terms what this diff represents!
func (ad *AppDiff) Describe() string {
	var buf bytes.Buffer
	if (ad.Was == nil) || (ad.Is == nil) {
		s := fmt.Sprintf("%v -> %v", ad.Was, ad.Is)
		return s
	}

	if ad.Is.DownloadURL != ad.Was.DownloadURL {
		buf.WriteString(fmt.Sprintf("   DownloadURL %s -> %s ", ad.Is.DownloadURL, ad.Was.DownloadURL))
	}
	if ad.Is.DownloadUser != ad.Was.DownloadUser {
		buf.WriteString(fmt.Sprintf("   DownloadUser %s -> %s ", ad.Is.DownloadUser, ad.Was.DownloadUser))
	}

	if ad.Is.DownloadPassword != ad.Was.DownloadPassword {
		buf.WriteString(fmt.Sprintf("    DownloadPassword %s -> %s ", ad.Is.DownloadPassword, ad.Was.DownloadPassword))
	}
	if ad.Is.Binary != ad.Was.Binary {
		buf.WriteString(fmt.Sprintf("    Binary %s -> %s ", ad.Is.Binary, ad.Was.Binary))
	}
	if ad.Is.BuildID != ad.Was.BuildID {
		buf.WriteString(fmt.Sprintf("    BuildID %d -> %d ", ad.Is.BuildID, ad.Was.BuildID))
	}
	if ad.Is.Instances != ad.Was.Instances {
		buf.WriteString(fmt.Sprintf("    Instances %d -> %d ", ad.Is.Instances, ad.Was.Instances))
	}
	if ad.Is.Public != ad.Was.Public {
		buf.WriteString(fmt.Sprintf("    Instances %d -> %d ", ad.Is.Instances, ad.Was.Instances))
	}
	// TODO: display difference in args
	if buf.String() == "" {
		buf.WriteString(fmt.Sprintf("Difference in args:\nad1=%v\nad2=%v\n", ad.Is, ad.Was))
	}
	return fmt.Sprintf(" Difference (%v) %s", ad.Was, buf.String())
}

// compare groupdefinition request and work out the differences
func Compare(def1, def2 *pb.GroupDefinitionRequest) (*Diff, error) {
	diff := &Diff{}

	if def1.Namespace != def2.Namespace {
		return nil, errors.New(fmt.Sprintf("Comparing two different namespaces makes no sense (%s and %s). Bug?", def1.Namespace, def2.Namespace))
	}
	// 1. find all applications that exist in def1 but not def2
	findNonExists(def1, def2, false, diff)
	// 2. find all applications that exist in def2 but not def1
	findNonExists(def2, def1, true, diff)
	// 3. find all applications that exist in def1 and def2 but are different
	fmt.Printf("Checking wether applications are identical...\n")
	for _, ad1 := range def1.Applications {
		ad2 := findSame(def2, ad1)
		if ad2 == nil {
			continue
		}
		// here are the comparison snowflakes...
		// if stuff isn't defined in the new group definition, then
		// we copy from existing group definition as a default
		// this is to avoid having to list a specific build id in a config file,
		// we want the user to conveniently call "updateapp" with a buildid instead
		// (the most common usecase)
		if ad2.BuildID == 0 {
			ad2.BuildID = ad1.BuildID
		}
		if ad2.Instances == 0 {
			ad2.Instances = ad1.Instances
		}
		// end snowflakes
		if IsIdentical(ad1, ad2) {
			continue
		}

		adiff := AppDiff{}
		adiff.Is = ad1
		adiff.Was = ad2
		diff.AppDiffs = append(diff.AppDiffs, adiff)
	}
	fmt.Printf("Found %d differences:\n", len(diff.AppDiffs))
	for _, x := range diff.AppDiffs {
		fmt.Printf("Diff: %s\n", x.Describe())
	}
	return diff, nil
}

// find application ad in def2 (based on a call to isSame)
func findSame(def *pb.GroupDefinitionRequest, ad *pb.ApplicationDefinition) *pb.ApplicationDefinition {
	for _, ad2 := range def.Applications {
		if isSame(ad2, ad) {
			return ad2
		}
	}
	return nil
}

// add all that exist in def1 but not in def2 to diff
func findNonExists(def1, def2 *pb.GroupDefinitionRequest, reverse bool, diff *Diff) error {
	for _, ad := range def1.Applications {
		if !doesExistInDef(ad, def2.Applications) {
			nad := AppDiff{}
			if reverse {
				nad.Is = ad
			} else {
				nad.Was = ad
			}
			diff.AppDiffs = append(diff.AppDiffs, nad)
		}
	}
	return nil
}

func doesExistInDef(ad *pb.ApplicationDefinition, ads []*pb.ApplicationDefinition) bool {
	for _, x := range ads {
		if isSame(ad, x) {
			return true
		}
	}
	return false
}

// returns true if both registration arrays are identical
func IsAutoRegistrationIdentical(ad1, ad2 *pb.ApplicationDefinition) bool {
	for _, x := range ad1.AutoRegs {
		if !doesAutoRegExist(x, ad2) {
			return false
		}
	}
	for _, x := range ad2.AutoRegs {
		if !doesAutoRegExist(x, ad1) {
			return false
		}
	}
	return true
}

func doesAutoRegExist(a *pb.AutoRegistration, ad *pb.ApplicationDefinition) bool {
	for _, x := range ad.AutoRegs {
		if (x.Portdef == a.Portdef) && (x.ServiceName == a.ServiceName) && (x.ApiTypes == a.ApiTypes) {
			return true
		}
	}
	return false
}

// returns true if both arg arrays are identical
func AreArgsIdentical(ad1, ad2 *pb.ApplicationDefinition) bool {
	// check args
	for _, a1 := range ad1.Args {
		if !isStringInArray(a1, ad2.Args) {
			return false
		}
	}
	for _, a2 := range ad2.Args {
		if !isStringInArray(a2, ad1.Args) {
			return false
		}
	}
	return true
}
func AppLimitsAreIdentical(ad1, ad2 *pb.ApplicationDefinition) bool {
	al1 := ad1.Limits
	al2 := ad2.Limits
	if al1.MaxMemory != al2.MaxMemory {
		return false
	}
	return true
}

// check if these two application definitions are identical
// (in terms of actual deployment)
func IsIdentical(ad1, ad2 *pb.ApplicationDefinition) bool {
	if (ad1.DownloadURL != ad2.DownloadURL) || (ad1.DownloadUser != ad2.DownloadUser) || (ad1.DownloadPassword != ad2.DownloadPassword) || (ad1.Binary != ad2.Binary) || (ad1.BuildID != ad2.BuildID) || (ad1.Instances != ad2.Instances) || (ad1.Machines != ad2.Machines) || (ad1.Critical != ad2.Critical) || (ad1.AlwaysOn != ad2.AlwaysOn) {
		return false
	}
	if ad1.StaticTargetDir != ad2.StaticTargetDir {
		return false
	}
	if !AppLimitsAreIdentical(ad1, ad2) {
		return false
	}
	if !AreArgsIdentical(ad1, ad2) {
		return false
	}
	if !IsAutoRegistrationIdentical(ad1, ad2) {
		return false
	}

	return true
}

// check if this is the same application (but not necessarily identical)
func isSame(ad1, ad2 *pb.ApplicationDefinition) bool {
	if ad1.RepositoryID != ad2.RepositoryID {
		return false
	}
	if ad1.Binary != ad2.Binary {
		return false
	}
	return true
}

// return true if we find a string in array
func isStringInArray(lookfor string, within []string) bool {
	for _, a := range within {
		if lookfor == a {
			return true
		}
	}
	return false
}
