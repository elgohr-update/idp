package identities

import (
  "net/http"
  "github.com/sirupsen/logrus"
  "github.com/gin-gonic/gin"

  "github.com/charmixer/idp/environment"
  "github.com/charmixer/idp/gateway/idp"
  . "github.com/charmixer/idp/models"
)

func PostDeleteVerification(env *environment.State, route environment.Route) gin.HandlerFunc {
  fn := func(c *gin.Context) {

    log := c.MustGet(environment.LogKey).(*logrus.Entry)
    log = log.WithFields(logrus.Fields{
      "func": "PostDeleteVerification",
    })

    var input IdentitiesDeleteVerificationRequest
    err := c.BindJSON(&input)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
      c.Abort()
      return
    }

    denyResponse := IdentitiesDeleteVerificationResponse{
      Id: input.Id,
      Verified: false,
      RedirectTo: "",
    }

    identity, exists, err := idp.FetchIdentityById(env.Driver, input.Id)
    if err != nil {
      log.Debug(err.Error())
      c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch Identity"})
      c.Abort();
      return
    }

    if exists == false {
      log.WithFields(logrus.Fields{"id": input.Id}).Debug("Identity not found")
      c.JSON(http.StatusNotFound, gin.H{"error": "Identity not found"})
      c.Abort();
      return
    }

    valid, err := idp.ValidatePassword(identity.OtpDeleteCode, input.VerificationCode)
    if err != nil {
      log.Debug(err.Error())
      log.WithFields(logrus.Fields{
        "id": denyResponse.Id,
        "verified": denyResponse.Verified,
        "redirect_to": denyResponse.RedirectTo,
      }).Debug("Delete verification rejected")
      c.JSON(http.StatusOK, denyResponse)
      c.Abort();
      return
    }

    if valid == true {

      log.WithFields(logrus.Fields{"fixme":1}).Debug("Revoke all access tokens for identity - put them on revoked list or rely on expire")
      log.WithFields(logrus.Fields{"fixme":1}).Debug("Revoke all consents in hydra for identity - this is probably aap?")

      n := idp.Identity{
        Id: identity.Id,
      }
      updatedIdentity, err := idp.DeleteIdentity(env.Driver, n)
      if err != nil {
        log.Debug(err.Error())
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete identitiy failed"})
        c.Abort();
        return
      }

      acceptResponse := IdentitiesDeleteVerificationResponse{
        Id: updatedIdentity.Id,
        Verified: true,
        RedirectTo: input.RedirectTo,
      }
      log.WithFields(logrus.Fields{
        "id": acceptResponse.Id,
        "verified": acceptResponse.Verified,
        "redirect_to": acceptResponse.RedirectTo,
      }).Debug("Identity deleted")
      c.JSON(http.StatusOK, acceptResponse)
      return
    }

    // Deny by default
    log.WithFields(logrus.Fields{
      "id": denyResponse.Id,
      "verified": denyResponse.Verified,
      "redirect_to": denyResponse.RedirectTo,
    }).Debug("Delete verification rejected")
    c.JSON(http.StatusOK, denyResponse)
  }
  return gin.HandlerFunc(fn)
}
