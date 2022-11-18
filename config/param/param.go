// Copyright 2016 Fake Twitter, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package param

// properties
const (
	Enabled          = "elon.enabled"
	Leashed          = "elon.leashed"
	ScheduleEnabled  = "elon.schedule_enabled"
	Accounts         = "elon.accounts"
	StartHour        = "elon.start_hour"
	EndHour          = "elon.end_hour"
	TimeZone         = "elon.time_zone"
	CronPath         = "elon.cron_path"
	TermPath         = "elon.term_path"
	TermAccount      = "elon.term_account"
	MaxTeams          = "elon.max_apps"
	Trackers         = "elon.trackers"
	ErrorCounter     = "elon.error_counter"
	Decryptor        = "elon.decryptor"
	OutageChecker    = "elon.outage_checker"
	CronExpression   = "elon.cron_expression"
	ScheduleCronPath = "elon.schedule_cron_path"
	SchedulePath     = "elon.schedule_path"
	LogPath          = "elon.log_path"

	// sysbreaker
	SysbreakerEndpoint          = "sysbreaker.endpoint"
	SysbreakerCertificate       = "sysbreaker.certificate"
	SysbreakerEncryptedPassword = "sysbreaker.encrypted_password"
	SysbreakerUser              = "sysbreaker.user"
	SysbreakerX509Cert          = "sysbreaker.x509_cert"
	SysbreakerX509Key           = "sysbreaker.x509_key"
	// database
	DatabaseHost              = "database.host"
	DatabasePort              = "database.port"
	DatabaseUser              = "database.user"
	DatabaseEncryptedPassword = "database.encrypted_password"
	DatabaseName              = "database.name"

	// dynamic property provider
	DynamicProvider = "dynamic.provider"
	DynamicEndpoint = "dynamic.endpoint"
	DynamicPath     = "dynamic.path"
)
