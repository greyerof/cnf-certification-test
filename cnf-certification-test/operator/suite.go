// Copyright (C) 2020-2021 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

package operator

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/common"
	"github.com/test-network-function/cnf-certification-test/cnf-certification-test/identifiers"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
)

//
// All actual test code belongs below here.  Utilities belong above.
//
var _ = ginkgo.Describe(common.OperatorTestKey, func() {
	var env provider.TestEnvironment
	ginkgo.BeforeEach(func() {
		provider.BuildTestEnvironment()
		env = provider.GetTestEnvironment()
	})

	testID := identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorInstallStatusSucceededIdentifier)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testOperatorInstallationPhaseSucceeded(&env)
	})

	testID = identifiers.XformToGinkgoItIdentifier(identifiers.TestOperatorNoPrivileges)
	ginkgo.It(testID, ginkgo.Label(testID), func() {
		testOperatorInstallationWithoutPrivileges(&env)
	})
})

func testOperatorInstallationPhaseSucceeded(env *provider.TestEnvironment) {
	badCsvs := []string{}
	if len(env.Csvs) == 0 {
		ginkgo.Skip("No CSVs to perform test, skipping.")
	}

	for _, csv := range env.Csvs {
		if csv.Status.Phase != v1alpha1.CSVPhaseSucceeded {
			badCsvs = append(badCsvs, fmt.Sprintf("%s.%s", csv.Namespace, csv.Name))
			tnf.ClaimFilePrintf("CSV %s (ns %s) is in phase %s. Expected phase is %s",
				csv.Name, csv.Namespace, csv.Status.Phase, v1alpha1.CSVPhaseSucceeded)
		}
	}

	if n := len(badCsvs); n > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d CSVs whose phase is not %s.", n, v1alpha1.CSVPhaseSucceeded))
	}
}

func testOperatorInstallationWithoutPrivileges(env *provider.TestEnvironment) {
	badCsvs := []string{}
	if len(env.Csvs) == 0 {
		ginkgo.Skip("No CSVs to perform test, skipping.")
	}

	for _, csv := range env.Csvs {
		clusterPermissions := csv.Spec.InstallStrategy.StrategySpec.ClusterPermissions
		if len(clusterPermissions) == 0 {
			logrus.Debugf("No clusterPermissions found in csv %s (ns %s)", csv.Name, csv.Namespace)
			continue
		}

		LOOP:
		for i := range clusterPermissions {
			permission := clusterPermissions[i]
			for ruleIndex := range permission.Rules {
				if n := len(permission.Rules[ruleIndex].ResourceNames); n > 0 {
					tnf.ClaimFilePrintf("CSV %s (ns %s) has cluster permissions on %d resource names.", csv.Name, csv.Namespace, n)
					badCsvs = append(badCsvs, fmt.Sprintf("%s.%s", csv.Namespace, csv.Name))
					break LOOP
				}
			}
		}
	}

	if n := len(badCsvs); n > 0 {
		ginkgo.Fail(fmt.Sprintf("Found %d CSVs with priviledges on some resource names.", n))
	}
}
