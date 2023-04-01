package envvars

import (
	"encoding/base64"
//	"errors"
	"fmt"
	"os"
//	"path"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

/*	Package `envvars` encapsulates common lambda function initialization tasks such as making sure
	all required environment variables are present and decrypting any that are encrypted.

Sample usage:
Suppose you have defined these environment variables in the AWS lambda function:
	Name				security
	----				--------
AWS
	awsLambdaFuncName	plaintext
	awsRegion			plaintext
SQL Server
	sqlHost				plaintext
	sqlUsername			plaintext
	sqlPassword			ENCRYPTED

Note:
	The reason for awsLambdaFuncName and awsRegion is that when encryption is used, the policy in
	the lambda function's execution role references its name, and the region is used to establish the KMS client.
	There's not a trivial way of acquiring these values programmatically, and if we hardcode them
	then we have to change the code and rebuild whenever we change the lambda function name or region in AWS.
	So making these environment variables seems a reasonable solution.

	The downside is that since KAwsRegion and kAwsLambdaFuncName are defined in this package, those 2
	environment variables as defined in the lambda function will have to match the spelling and casing exactly.

Init()
	const (
		kSqlHost		= `sqlHost`
		kSqlUsername	= `sqlUsername`
		kSqlPassword	= `sqlPassword`
	)

	var err error

	mEnvVars := envvars.NewEnvVarMap()

	if err = mEnvVars.Add(kSqlHost, false); nil != err {
		log.Panicln(err)
	}
	if err = mEnvVars.Add(kSqlUsername, false); nil != err {
		log.Panicln(err)
	}
	if err = mEnvVars.Add(kSqlPassword, true); nil != err {
		log.Panicln(err)
	}

	//	mEnvVars.Validate() will automatically add awsLambdaFuncName and awsRegion if any of the above Adds have encrypted = true.
	if err = mEnvVars.Validate(); nil != err {
		log.Panicln(err)
	}

	//	populate the config struct to pass into whatever package's Init()
	pConfigSqlServer := &sqlserver.TConfig{
		Host:		mEnvVars[kSqlHost].Plaintext,
		Port:		1433,
		Username:	mEnvVars[kSqlUsername].Plaintext,
		Password:	mEnvVars[kSqlPassword].Plaintext,
		Database:	`imt`,
		Verbose:	false,
	}

	if err = sqlserver.Init(pConfigSqlServer); nil != err {
		log.Panicf(`sqlserver.Init() returned error: %v`, err)
	}
}//init()
*/

//\\//	package-scope constants and variables

const (
	KAwsRegion			= `awsRegion`
	kAwsLambdaFuncName	= `awsLambdaFuncName`
)


//\\//	type definitions (and attached methods)

type tEnvVar struct {
	Plaintext	string
	Ciphertext	string
//	Required	bool	//	so far everything is required
	Encrypted	bool
}

func (p *tEnvVar) decrypt(key string, pKMS *kms.KMS, encryptionContext aws.StringMap) (err error) {
	//func (enc *Encoding) DecodeString(s string) ([]byte, error)
	var xDecodedBytes []byte
	if xDecodedBytes, err = base64.StdEncoding.DecodeString(p.Ciphertext); nil == err {
		pDecryptInput := &kms.DecryptInput {
			CiphertextBlob: xDecodedBytes,
			EncryptionContext: encryptionContext,
		}

		//func (c *KMS) Decrypt(input *DecryptInput) (*DecryptOutput, error)
		var pDecryptOutput *kms.DecryptOutput
		if pDecryptOutput, err = pKMS.Decrypt(pDecryptInput); nil == err {
			//	Plaintext is a byte array, so convert to string
			p.Plaintext = string(pDecryptOutput.Plaintext[:])
		} else {
			err = fmt.Errorf(`Failed to decrypt %s environment variable: %w`, key, err)
		}
	} else {
		err = fmt.Errorf(`Failed to base64-decode %s environment variable: %w`, key, err)
	}

	return
}


type TEnvVarMap map[string]*tEnvVar

func (m TEnvVarMap) Add(key string, /*required, */encrypted bool) (err error) {
	if _, present := m[key]; present {
		err = fmt.Errorf(`Key "%s" repeated`, key)
	} else {
		m[key] = newEnvVar(key, encrypted)
	}

	return
}

func (m TEnvVarMap) addIfNotPresent(key string, /*required, */encrypted bool) {
	if _, present := m[key]; !present {
		m[key] = newEnvVar(key, encrypted)
	}

	return
}

func (m TEnvVarMap) Validate() (err error) {
	var needsEncryption bool

	//	first determine if encryption is needed
	for _, pEnvVar := range m {
		if pEnvVar.Encrypted {
			needsEncryption = true
			break
		}
	}

	//	if so, then Add KAwsRegion and kAwsLambdaFuncName
	if needsEncryption {
		m.addIfNotPresent(kAwsLambdaFuncName, false)
		m.addIfNotPresent(KAwsRegion, false)
	}

	xMissing := make([]string, 0, len(m))
	//	range over the map to confirm that the gang is all here
	for key, pEnvVar := range m {
		if pEnvVar.Encrypted {
			if 0 == len(pEnvVar.Ciphertext) {
				xMissing = append(xMissing, key)
			}
		} else {
			if 0 == len(pEnvVar.Plaintext) {
				xMissing = append(xMissing, key)
			}
		}
	}

	if 0 == len(xMissing) {
		if needsEncryption {
			//	decrypt the encrypted environment variables

			//	yeschiree there's a seschschion in seschschion
			pSession		:= session.Must(session.NewSession())
			pConfigRegion	:= aws.NewConfig().WithRegion(m[KAwsRegion].Plaintext)

			//func New(p client.ConfigProvider, cfgs ...*aws.Config) *KMS
			//	Create a KMS client from just a session.
//			pKMS			:= kms.New(pSession)
			//	Create a KMS client with additional configuration
			pKMS			:= kms.New(pSession, pConfigRegion)

			encryptionContext := aws.StringMap(map[string]string{`LambdaFunctionName`: m[kAwsLambdaFuncName].Plaintext})

			//	range over the map to decrypt the items requiring it
			for key, pEnvVar := range m {
				if pEnvVar.Encrypted {
					if err = pEnvVar.decrypt(key, pKMS, encryptionContext); nil != err {
						break
					}
				}
			}
		}//needsEncryption
	} else {
		err = fmt.Errorf(`Missing environemnt variables: %s`, strings.Join(xMissing, `, `))
	}

	return
}


//\\//	functions

func newEnvVar(key string, /*required, */encrypted bool) *tEnvVar {
	//	evaluate os.Getenv(key) and place it in either Plaintext or Ciphertext depending on encrypted
	pEnvVar	:= new(tEnvVar)
	value	:= os.Getenv(key)

	if encrypted {
		pEnvVar.Ciphertext	= value
		pEnvVar.Encrypted	= true
	} else {
		pEnvVar.Plaintext	= value
	}

	return pEnvVar
}

func NewEnvVarMap() TEnvVarMap {
	return make(TEnvVarMap)
}


