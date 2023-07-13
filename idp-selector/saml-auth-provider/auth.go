package main

import (
	"fmt"
	"net/http"
	"strings"

	"maverics/auth"
	"maverics/log"
	"maverics/session"
)

const (
	// oldcoIDP represents the name of the IDP that oldco users authenticate against.
	oldcoIDP = "Okta"
	// newcoIDP represents the name of the IDP that newco users authenticate
	// against.
	newcoIDP = "Azure"
	// newcoUserSuffix is used to distinguish newco users from oldco users.
	newcoUserSuffix = "@newco.com"
)

// IsAuthenticated determines if the user is authenticated. Authentication status is
// derived by querying the session cache.
func IsAuthenticated(_ *auth.SAMLProvider, _ http.ResponseWriter, req *http.Request) bool {
	log.Debug("se", "determining if user is authenticated")

	return isAuthenticated(req)
}

func isAuthenticated(req *http.Request) bool {
	if session.GetString(req, fmt.Sprintf("%s.authenticated", oldcoIDP)) == "true" {
		log.Debug("se", "user is authenticated", "idpName", oldcoIDP)
		return true
	}
	if session.GetString(req, fmt.Sprintf("%s.authenticated", newcoIDP)) == "true" {
		log.Debug("se", "user is authenticated", "idpName", newcoIDP)
		return true
	}
	log.Debug("se", "user is not authenticated")
	return false
}

// Authenticate authenticates the user against the IDP that they select.
func Authenticate(sp *auth.SAMLProvider, rw http.ResponseWriter, req *http.Request) {

	// Render the IDP picker if the user has not yet been authenticated and if they
	// have not entered a value in the idp picker form. We are passing the original
	// SAMLRequest as a hidden form field in the IDP Picker form so that upon
	// submitting the idp form, it comes back through the SAML Auth Provider as a
	// valid SAML request.
	hasIDPBeenPicked := req.FormValue("username")
	idpForm := fmt.Sprintf(idpFormTemplate, req.FormValue("SAMLRequest"))
	if !isAuthenticated(req) && len(hasIDPBeenPicked) == 0 {
		log.Debug("se", "rendering idp picker")
		rw.Write([]byte(idpForm))
		return
	}

	log.Info("se", "authenticating user")

	log.Info("se", "parsing IDP Picker POST request")
	err := req.ParseForm()
	if err != nil {
		log.Error("se", fmt.Sprintf(
			"failed to parse form from request: %s",
			err,
		))
		http.Error(
			rw,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	var (
		username     = req.Form.Get("username")
		employeeType = "oldco"
		idpName      = oldcoIDP
	)
	if strings.HasSuffix(username, newcoUserSuffix) {
		employeeType = "newco"
		idpName = newcoIDP
	}

	idp, found := sp.IDPs[idpName]
	if !found {
		log.Error("se", "idp not found", "IDPName", idpName)
		http.Error(
			rw,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	log.Info(
		"se", "authenticating user against idp",
		"IDPName", idpName,
		"username", username,
		"employeeType", employeeType,
	)

	idp.CreateRequest().Login(rw, req)
}

// BuildClaims builds attributes will be shared to the up-stream SAML Service
// Provider as a SAML 2.0 AttributeStatement.
func BuildClaims(
	_ *auth.SAMLProvider,
	_ http.ResponseWriter,
	req *http.Request,
) (map[string]string, error) {
	log.Debug("se", "building claims for ID token")

	var (
		selectedIDP   string
		claimsMapping = map[string]string{
			"email":       "email",
			"name":        "email",
			"family_name": "family_name",
			"given_name":  "given_name",
			"middle_name": "middle_name",
			"nickname":    "nickname",
			"picture":     "picture",
			"updated_at":  "updated_at",
		}
		returnClaims = make(map[string]string)
	)

	if session.GetString(req, fmt.Sprintf("%s.authenticated", oldcoIDP)) == "true" {
		selectedIDP = oldcoIDP
	}
	if session.GetString(req, fmt.Sprintf("%s.authenticated", newcoIDP)) == "true" {
		selectedIDP = newcoIDP
	}
	if selectedIDP == "" {
		return nil, fmt.Errorf("unable to determine which IDP to retrieve claims from")
	}

	for claim, mapping := range claimsMapping {
		claimValue := session.GetString(req, fmt.Sprintf("%s.%s", selectedIDP, mapping))
		if claimValue == "" {
			continue
		}
		returnClaims[claim] = claimValue
	}

	return returnClaims, nil
}

// idpFormTemplate is a basic form that is rendered in order to enable the user to pick which
// IDP they want to authenticate against. The markup can be styled as necessary,
// loaded from an external file, be rendered as a dynamic template, etc.
const idpFormTemplate = `
<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="ie=edge">
  <title>Let us log you in</title>
  <link rel="shortcut icon" href = "https://launchpad.stratademo.com/web/image/favicon.png" type="image/x-icon"/>
  <!-- Bootstrap , fonts & icons  -->
  <link rel="stylesheet" type="text/css" href = "https://launchpad.stratademo.com/web/css/bootstrap.css"/>
  <link rel="stylesheet" type="text/css" href = "https://launchpad.stratademo.com/web/css/style.css"/>
  <link rel="stylesheet" type="text/css" href = "https://launchpad.stratademo.com/web/fonts/typography-font/typo.css"/>
  <link rel="stylesheet" type="text/css" href = "https://launchpad.stratademo.com/web/fonts/fontawesome-5/css/all.css"/>
  <link href="https://fonts.googleapis.com/css2?family=Karla:wght@300;400;500;600;700;800&display=swap" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css2?family=Gothic+A1:wght@400;500;700;900&display=swap" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css2?family=Work+Sans:wght@400;500;600;700;800;900&display=swap" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css2?family=Rubik:wght@400;500;600;700;800;900&display=swap" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800;900&display=swap" rel="stylesheet">
  <link href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;600;700;800;900&display=swap" rel="stylesheet">
  <!-- Plugin'stylesheets  -->
  <link rel="stylesheet" type="text/css" href="https://launchpad.stratademo.com/web/plugins/aos/aos.min.css">
  <link rel="stylesheet" type="text/css" href="https://launchpad.stratademo.com/web/plugins/fancybox/jquery.fancybox.min.css">
  <link rel="stylesheet" type="text/css" href="https://launchpad.stratademo.com/web/lugins/nice-select/nice-select.min.css">
  <link rel="stylesheet" type="text/css" href="https://launchpad.stratademo.com/web/plugins/slick/slick.min.css">
  <!-- Vendor stylesheets  -->
  <link rel="stylesheet" type="text/css" href="https://launchpad.stratademo.com/web/css/main.css">
  <!-- Custom stylesheet -->
</head>

<body data-theme-mode-panel-active data-theme="light" style="font-family: 'Mazzard H';">
  <div class="site-wrapper overflow-hidden position-relative">
    <!-- Site Header -->
    <!-- Preloader -->
    <!-- <div id="loading">
    <div class="preloader">
     <img src="https://launchpad.stratademo.com/web/image/preloader.gif" alt="preloader">
   </div>
   </div>    -->
    <!--Site Header Area -->
    <header class="site-header site-header--menu-right sign-in-menu-1 site-header--absolute site-header--sticky">
      <div class="container">
        <nav class="navbar site-navbar">
          <!-- Brand Logo-->
          <div class="brand-logo">
            <a href="#">
            </a>
          </div>
          <div class="menu-block-wrapper">
            <div class="menu-overlay"></div>
          </div>
          <div class="header-btn sign-in-header-btn-1 ms-auto d-none d-xs-inline-flex">
          </div>
          <div class="header-btns  ">
            <a class="" href="#">
            </a>
            <a class="" href="#">
            </a>
          </div>
          <!-- mobile menu trigger -->
          <div class="mobile-menu-trigger">
            <span></span>
          </div>
          <!--/.Mobile Menu Hamburger Ends-->
        </nav>
      </div>
    </header>
    <!-- navbar- -->
    <!-- Sign In Area -->
    <div class="reset-password-1">
      <div class="container">
        <div class="row justify-content-lg-end justify-content-center">
          <div class="col-xl-5 col-lg-6 col-md-8">
            <div class="reset-password-1-box  justify-content-lg-end">
              <div class="heading text-center">
                <h2>Provide username</h2>
                <p>This lets us know where to send you for authentication</p>
              </div>
              <form method="post">
                <div class="form-group">
                  <label>Email</label>
                  <input type="hidden" name="SAMLRequest" id="SAMLRequest" value="%s">
                  <input type="text" name="username" id="username" class="form-control" placeholder="ex: jdoe@newco.com">
                </div>
                  <center><input type="submit">

                <div class="create-new-acc-text text-center">
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>
    </div>
    <!--Footer Area-->
    <footer class="footer-sign-in-1">
      <div class="container">
        <div class="row">
          <div class="col-lg-5 col-sm-9">
            <div class="row">
              <div class="col-xl-9 col-lg-10 col-md-8">
                <a href="#"><img src="https://launchpad.stratademo.com/web/image/logos/logo-paste.png" alt="" class="footer-logo"></a>
                <div class="content">
                </div>
                <div class="social-icons">
                </div>
              </div>
            </div>
          </div>
          <div class="col-lg-7 col-md-12">
            <div class="row">
              <div class="col-lg-3 col-md-3 col-sm-4 col-xs-6">
                <div class="footer-widget">
                </div>
              </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </footer>
  </div>
</body>

</html>
`
