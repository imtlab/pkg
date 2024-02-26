package envvars

import (
	"encoding/base64"
	"fmt"
//	"log"	//	only used if lines marked DEBUG are uncommented
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

/*
###	Description:
Package `envvars` encapsulates common lambda function initialization tasks such as making sure
all required environment variables are present and decrypting any that are encrypted.

###	Notes:
The AWS lambda runtime sets several environment variables during initialization that this package references
if Encrypted = true for any of the tEnvVars added to the TEnvVarMap:
	AWS_LAMBDA_FUNCTION_NAME
	AWS_DEFAULT_REGION	The default AWS Region where the Lambda function is executed.
	AWS_REGION			The AWS Region where the Lambda function is executed. If defined, this value overrides AWS_DEFAULT_REGION.
(See https://docs.aws.amazon.com/lambda/latest/dg/configuration-envvars.html#configuration-envvars-runtime)

### Sample usage:
Suppose you have defined these environment variables in the AWS lambda function:
Name		security
----		--------
sqlHost		plaintext
sqlUsername	plaintext
sqlPassword	ENCRYPTED

In the init() of your main package:
init() {
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
}
*/

//\\//	package-scope constants and variables


//\\//	type definitions (and attached methods)

type tEnvVar struct {
	Plaintext	string
	Ciphertext	string
//	Required	bool	//	so far everything is required
	Encrypted	bool
}

func (p *tEnvVar) decrypt(key string, pKMS *kms.KMS, encryptionContext map[string]*string) (err error) {
	//func (enc *Encoding) DecodeString(s string) ([]byte, error)
	var xDecodedBytes []byte
	if xDecodedBytes, err = base64.StdEncoding.DecodeString(p.Ciphertext); nil == err {
		/*	NOTE:
			According to the kms package documentation at https://pkg.go.dev/github.com/aws/aws-sdk-go/service/kms#DecryptInput
			EncryptionContext is "optional, but it is strongly recommended".
			But it says essentially the same thing about the KeyId struct field.  To Wit:
				If you used a symmetric encryption KMS key, KMS can get the KMS key from metadata
				that it adds to the symmetric ciphertext blob.  However, it is always recommended
				as a best practice.  This practice ensures that you use the KMS key that you intend.
			Yet the sample code in the "Decrypt secrets snippet" (provided by the AWS Lamba Function
			configuration UI when encrypting the environment variables) doesn't include KeyId.
		*/
		pDecryptInput := &kms.DecryptInput{
			CiphertextBlob:		xDecodedBytes,
			EncryptionContext:	encryptionContext,
		}

//		log.Println(`Calling pKMS.Decrypt()`)											//<<<<	DEBUG

		//func (c *KMS) Decrypt(input *DecryptInput) (*DecryptOutput, error)
		var pDecryptOutput *kms.DecryptOutput
		if pDecryptOutput, err = pKMS.Decrypt(pDecryptInput); nil == err {
//		log.Println(`Exited pKMS.Decrypt()`)											//<<<<	DEBUG

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

func (m TEnvVarMap) Validate() (err error) {
	const (
		kAwsLambdaFuncName	= `AWS_LAMBDA_FUNCTION_NAME`
		kAwsDefaultRegion	= `AWS_DEFAULT_REGION`	//	presumably always populated?
		kAwsRegion			= `AWS_REGION`			//	presumably optional?
	)

	var needsEncryption bool

	//	first determine if encryption is needed
	for _, pEnvVar := range m {
		if pEnvVar.Encrypted {
			needsEncryption = true
			break
		}
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

			//	We'll need awsRegion for our session, and awsLambdaFuncName for the encryptionContext.
			var awsRegion, awsLambdaFuncName string

			if awsRegion = os.Getenv(kAwsRegion); 0 == len(awsRegion) {
				if awsRegion = os.Getenv(kAwsDefaultRegion); 0 == len(awsRegion) {
					xMissing = append(xMissing, kAwsDefaultRegion)
				}
			}

			if awsLambdaFuncName = os.Getenv(kAwsLambdaFuncName); 0 == len(awsRegion) {
				xMissing = append(xMissing, kAwsLambdaFuncName)
			}

			if 0 == len(xMissing) {
				//	yeschiree there's a seschschion in seschschion
				pSession		:= session.Must(session.NewSession())
				pConfigRegion	:= aws.NewConfig().WithRegion(awsRegion)
				//func New(p client.ConfigProvider, cfgs ...*aws.Config) *KMS
				//	Create a KMS client with additional configuration
				pKMS			:= kms.New(pSession, pConfigRegion)

				encryptionContext := aws.StringMap(map[string]string{`LambdaFunctionName`: awsLambdaFuncName})

				//	range over the map to decrypt the items requiring it
				for key, pEnvVar := range m {
//					log.Printf(`key = %s; Encrypted = %v`, key, pEnvVar.Encrypted)		//<<<<	DEBUG
					if pEnvVar.Encrypted {
						if err = pEnvVar.decrypt(key, pKMS, encryptionContext); nil != err {
							break
						}
					}
				}
			} else {
				err = fmt.Errorf(`Missing lambda runtime environemnt variables: %s`, strings.Join(xMissing, `, `))
			}
		}//needsEncryption
	} else {
		err = fmt.Errorf(`Missing configured environemnt variables: %s`, strings.Join(xMissing, `, `))
	}

	return
}//Validate()


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


