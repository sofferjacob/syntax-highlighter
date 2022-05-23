package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jkomyno/nanoid"
	"github.com/sofferjacob/dashboard_api/db"
	"github.com/sofferjacob/dashboard_api/ent"
	"github.com/sofferjacob/dashboard_api/ent/googleoauthstate"
	"github.com/sofferjacob/dashboard_api/ent/project"
	"github.com/sofferjacob/dashboard_api/ent/source"
	"github.com/sofferjacob/dashboard_api/tokens"
)

func GetProjectSources(c *gin.Context) {
	claimsObj, ok := c.Get("project-claims")
	if !ok {
		c.JSON(400, gin.H{"error": "missing claims"})
		return
	}
	claims, ok := claimsObj.(*tokens.ProjectToken)
	if !ok {
		c.JSON(400, gin.H{"error": "invalid claims"})
		return
	}
	sources, err := db.Client.Source.Query().Where(source.SourceProject(claims.ProjectId)).All(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok", "sources": sources})
}

func DeleteSource(c *gin.Context) {
	claimsObj, ok := c.Get("project-claims")
	if !ok {
		c.JSON(400, gin.H{"error": "missing claims"})
		return
	}
	claims, ok := claimsObj.(*tokens.ProjectToken)
	if !ok {
		c.JSON(400, gin.H{"error": "invalid claims"})
		return
	}
	sourceId := c.Param("id")
	if sourceId == "" {
		c.JSON(400, gin.H{"error": "missing required param id"})
		return
	}
	source, err := db.Client.Source.Query().Where(source.And(source.SourceProject(claims.ProjectId), source.ID(sourceId))).Only(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "could not fetch source", "message": err.Error()})
		return
	}
	err = db.Client.Source.DeleteOne(source).Exec(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "could not delete source", "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

type CreateFacebookSourceParams struct {
	AccountId   string `json:"accountId" binding:"required"`
	AccessToken string `json:"accessToken" binding:"required"`
	SourceName  string `json:"sourceName" binding:"required"`
}

type FBTokenResponse struct {
	Token string `json:"access_token" binding:"required"`
}

func CreateFacebookSource(c *gin.Context) {
	claimsObj, ok := c.Get("project-claims")
	if !ok {
		c.JSON(400, gin.H{"error": "missing claims"})
		return
	}
	claims, ok := claimsObj.(*tokens.ProjectToken)
	if !ok {
		c.JSON(400, gin.H{"error": "invalid claims"})
		return
	}
	var params CreateFacebookSourceParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	urlStr := fmt.Sprintf("https://graph.facebook.com/oauth/access_token?grant_type=fb_exchange_token&client_id=%v&client_secret=%v&fb_exchange_token=%v", os.Getenv("FB_APP_ID"), os.Getenv("FB_APP_SECRET"), params.AccessToken)
	// fmt.Println(urlStr)
	reqUrl, _ := url.Parse(urlStr)
	req := &http.Request{
		Method: "GET",
		URL:    reqUrl,
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not trade token", "message": err.Error()})
		return
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not read FB response", "message": err.Error()})
		return
	}
	if response.StatusCode != 200 {
		c.JSON(500, gin.H{"error": fmt.Sprintf("could not trade token, invalid status code: %v response: %v", response.StatusCode, string(responseData))})
		return
	}
	// fmt.Println(string(responseData))
	var token FBTokenResponse
	err = json.Unmarshal(responseData, &token)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not parse FB response", "message": err.Error()})
		return
	}
	sourceId, _ := nanoid.Nanoid(10)
	source, err := db.Client.Source.Create().SetID(sourceId).SetType("fb").SetProjectID(claims.ProjectId).SetConfig(gin.H{
		"accountId":   params.AccountId,
		"accessToken": token.Token,
	}).SetName(params.SourceName).Save(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "could not create source", "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok", "source": source})
}

type CreateGoogleSourceParams struct {
	State string `json:"state" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type GoogleResponse struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func CreateGoogleSource(c *gin.Context) {
	claimsObj, ok := c.Get("project-claims")
	if !ok {
		c.JSON(400, gin.H{"error": "missing claims"})
		return
	}
	claims, ok := claimsObj.(*tokens.ProjectToken)
	if !ok {
		c.JSON(400, gin.H{"error": "invalid claims"})
		return
	}
	var params CreateGoogleSourceParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	urlStr := fmt.Sprintf(
		"https://oauth2.googleapis.com/token?code=%v&client_id=%v&client_secret=%v&grant_type=authorization_code&redirect_uri=%v",
		params.Code,
		os.Getenv("GA_CLIENT_ID"),
		os.Getenv("GA_SECRET"),
		os.Getenv("GA_REDIRECT_URI"),
	)
	reqUrl, _ := url.Parse(urlStr)
	req := &http.Request{
		Method: "POST",
		URL:    reqUrl,
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not trade code", "message": err.Error()})
		return
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not read GOOGLE response", "message": err.Error()})
		return
	}
	if response.StatusCode != 200 {
		c.JSON(500, gin.H{"error": fmt.Sprintf("could not trade code, invalid status code: %v response: %v", response.StatusCode, string(responseData))})
		return
	}
	var token GoogleResponse
	err = json.Unmarshal(responseData, &token)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not parse google response", "message": err.Error()})
		return
	}
	state, err := db.Client.GoogleOauthState.Query().Where(googleoauthstate.ID(params.State)).Only(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "could not fetch state", "message": err.Error()})
		return
	}
	_ = db.Client.GoogleOauthState.DeleteOne(state).Exec(context.Background())
	sourceId, _ := nanoid.Nanoid(10)
	config := gin.H{
		"refresh_token":       token.RefreshToken,
		"customer_id":         state.CustomerID,
		"manager_customer_id": state.ManagerID,
	}
	source, err := db.Client.Source.Create().
		SetID(sourceId).
		SetType("ga").
		SetProjectID(claims.ProjectId).
		SetConfig(config).
		SetName(state.SourceName).
		Save(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "could not create source", "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok", "source": source})
}

type CreateCPSourceParams struct {
	SourceName string `json:"source_name" binding:"required"`
	Numbers    []struct {
		Name   string `json:"name" binding:"required"`
		Number string `json:"number" binding:"required"`
	} `json:"numbers" binding:"required"`
}

func CreateCPSource(c *gin.Context) {
	claimsObj, ok := c.Get("project-claims")
	if !ok {
		c.JSON(400, gin.H{"error": "missing claims"})
		return
	}
	claims, ok := claimsObj.(*tokens.ProjectToken)
	if !ok {
		c.JSON(400, gin.H{"error": "invalid claims"})
		return
	}
	var params CreateCPSourceParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	sourceId, _ := nanoid.Nanoid(10)
	source, err := db.Client.Source.Create().
		SetName(params.SourceName).
		SetID(sourceId).
		SetType("cp").
		SetProjectID(claims.ProjectId).
		SetConfig(gin.H{
			"numbers": params.Numbers,
		}).Save(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "could not create source", "message": err.Error()})
	}
	c.JSON(200, gin.H{"status": "ok", "source": source, "url": "https://dash.gbs-digital.com/t/cp"})
}

func notifyCpHook(body db.CallPickerRecord) {
	data, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Could not post to CP hook: %v\n", err.Error())
		return
	}
	_, err = http.Post("https://hooks.zapier.com/hooks/catch/6820014/bz8ghqa", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Could not post to CP hook: %v\n", err.Error())
		return
	}
}

func CPTransport(c *gin.Context) {
	var body db.CallPickerRecord
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	fmt.Println(body)
	go notifyCpHook(body)
	projectId, err := db.Transport.GetCPSourceProject(body.CpNumber)
	if err != nil {
		c.JSON(500, gin.H{"error": "The project does not exist, is inactive or is not connected to Call Picker", "message": err.Error()})
		fmt.Printf("Error: source not found: %v\n", err.Error())
		return
	}
	err = db.Transport.InsertCpRecord(projectId, body)
	if err != nil {
		c.JSON(500, gin.H{"error": "could not insert record", "message": err.Error()})
		fmt.Printf("Error: could not insert record: %v\n", err.Error())
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func parseFbSource(v *ent.Source, sourcesObj gin.H) {
	sourcesObj[fmt.Sprintf("%v_%v_%v", v.Name, v.Type, v.SourceProject)] = gin.H{
		"type":         "facebook_marketing",
		"destinations": []string{fmt.Sprintf("postgres_%v", v.SourceProject)},
		"config": gin.H{
			"account_id":   v.Config["accountId"],
			"access_token": v.Config["accessToken"],
		},
		"collections": []gin.H{
			{
				"name":     "ads",
				"type":     "ads",
				"level":    "ad",
				"schedule": "*/20 * * * *",
				"parameters": gin.H{
					"fields": []string{"bid_amount", "adlabels", "creative", "status", "created_time", "updated_time", "targeting", "effective_status", "campaign_id", "adset_id", "conversion_specs", "recommendations", "id", "bid_info", "tracking_specs", "bid_type", "name", "account_id", "source_ad_id"},
				},
			},
			{
				"name":     "adset",
				"type":     "ads",
				"level":    "adset",
				"schedule": "*/20 * * * *",
				"parameters": gin.H{
					"fields": []string{"bid_amount", "adlabels", "creative", "status", "created_time", "updated_time", "targeting", "effective_status", "campaign_id", "adset_id", "conversion_specs", "recommendations", "id", "bid_info", "tracking_specs", "bid_type", "name", "account_id", "source_ad_id"},
				},
			},
			{
				"name":     "campaign",
				"type":     "ads",
				"level":    "campaign",
				"schedule": "*/30 * * * *",
				"parameters": gin.H{
					"fields": []string{"bid_amount", "adlabels", "creative", "status", "created_time", "updated_time", "targeting", "effective_status", "campaign_id", "adset_id", "conversion_specs", "recommendations", "id", "bid_info", "tracking_specs", "bid_type", "name", "account_id", "source_ad_id"},
				},
			},
			{
				"name":     "account",
				"type":     "account",
				"level":    "account",
				"schedule": "*/60 * * * *",
				"parameters": gin.H{
					"fields": []string{"bid_amount", "adlabels", "creative", "status", "created_time", "updated_time", "targeting", "effective_status", "campaign_id", "adset_id", "conversion_specs", "recommendations", "id", "bid_info", "tracking_specs", "bid_type", "name", "account_id", "source_ad_id"},
				},
			},
		},
	}
}

func parseGaSource(v *ent.Source, obj gin.H) {
	obj[fmt.Sprintf("%v_%v_%v", v.Name, v.Type, v.SourceProject)] = gin.H{
		"type":         "google_ads",
		"destinations": []string{fmt.Sprintf("postgres_%v", v.SourceProject)},
		"config": gin.H{
			"customer_id":         v.Config["customer_id"],
			"manager_customer_id": v.Config["manager_customer_id"],
			"auth": gin.H{
				"type":          "OAuth",
				"client_id":     os.Getenv("GA_CLIENT_ID"),
				"client_secret": os.Getenv("GA_SECRET"),
				"refresh_token": v.Config["refresh_token"],
			},
		},
		"schedule": "*/60 * * * *",
		"collections": []gin.H{
			{
				"name":       "accessible_bidding_strategy",
				"type":       "accessible_bidding_strategy",
				"table_name": "gads_accessible_bidding_strategy",
				"schedule":   "*/60 * * * *",
				"parameters": gin.H{
					"fields":     "accessible_bidding_strategy.id, accessible_bidding_strategy.maximize_conversion_value.target_roas, accessible_bidding_strategy.maximize_conversions.target_cpa, accessible_bidding_strategy.name, accessible_bidding_strategy.owner_customer_id, accessible_bidding_strategy.owner_descriptive_name, accessible_bidding_strategy.resource_name, accessible_bidding_strategy.target_cpa.target_cpa_micros, accessible_bidding_strategy.target_impression_share.cpc_bid_ceiling_micros, accessible_bidding_strategy.target_impression_share.location, accessible_bidding_strategy.target_impression_share.location_fraction_micros, accessible_bidding_strategy.target_roas.target_roas, accessible_bidding_strategy.target_spend.cpc_bid_ceiling_micros, accessible_bidding_strategy.target_spend.target_spend_micros, accessible_bidding_strategy.type",
					"start_date": "2021-01-01",
				},
			},
			{
				"name":       "account_budget",
				"type":       "account_budget",
				"table_name": "gads_account_budget",
				"schedule":   "*/60 * * * *",
				"parameters": gin.H{
					"fields":     "account_budget.adjusted_spending_limit_micros, account_budget.adjusted_spending_limit_type, account_budget.amount_served_micros, account_budget.approved_end_date_time, account_budget.approved_end_time_type, account_budget.approved_spending_limit_micros, account_budget.approved_spending_limit_type, account_budget.approved_start_date_time, account_budget.billing_setup, account_budget.id, account_budget.name, account_budget.notes, account_budget.pending_proposal.account_budget_proposal, account_budget.pending_proposal.creation_date_time, account_budget.pending_proposal.end_date_time, account_budget.pending_proposal.end_time_type, account_budget.pending_proposal.name, account_budget.pending_proposal.notes, account_budget.pending_proposal.proposal_type, account_budget.pending_proposal.purchase_order_number, account_budget.pending_proposal.spending_limit_micros, account_budget.pending_proposal.spending_limit_type, account_budget.pending_proposal.start_date_time, account_budget.proposed_end_date_time, account_budget.proposed_end_time_type, account_budget.proposed_spending_limit_micros, account_budget.proposed_spending_limit_type, account_budget.proposed_start_date_time, account_budget.purchase_order_number, account_budget.resource_name, account_budget.status, account_budget.total_adjustments_micros",
					"start_date": "2021-01-01",
				},
			},
			{
				"name":       "account_budget_proposal",
				"type":       "account_budget_proposal",
				"table_name": "gads_account_budget_proposal",
				"schedule":   "*/60 * * * *",
				"parameters": gin.H{
					"fields":     "account_budget_proposal.account_budget, account_budget_proposal.approval_date_time, account_budget_proposal.approved_end_date_time, account_budget_proposal.approved_end_time_type, account_budget_proposal.approved_spending_limit_micros, account_budget_proposal.approved_spending_limit_type, account_budget_proposal.approved_start_date_time, account_budget_proposal.billing_setup, account_budget_proposal.creation_date_time, account_budget_proposal.id, account_budget_proposal.proposal_type, account_budget_proposal.proposed_end_date_time, account_budget_proposal.proposed_end_time_type, account_budget_proposal.proposed_name, account_budget_proposal.proposed_notes, account_budget_proposal.proposed_purchase_order_number, account_budget_proposal.proposed_spending_limit_micros, account_budget_proposal.proposed_spending_limit_type, account_budget_proposal.proposed_start_date_time, account_budget_proposal.resource_name, account_budget_proposal.status",
					"start_date": "2021-01-01",
				},
			},
		},
	}
}

func GetSources(c *gin.Context) {
	apiKey := c.Param("key")
	if apiKey != os.Getenv("JITSU_KEY") {
		c.JSON(403, gin.H{"error": "Forbidden"})
		return
	}
	sources, err := db.Client.Source.Query().Where(source.HasProjectWith(project.Active(true))).All(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
	}
	sourcesObj := gin.H{}
	for _, v := range sources {
		if v.Type == "fb" {
			parseFbSource(v, sourcesObj)
		}
		if v.Type == "ga" {
			parseGaSource(v, sourcesObj)
		}
	}
	c.JSON(200, gin.H{"sources": sourcesObj})
}

type GoogleOauthParams struct {
	CustomerId string `json:"customerId" binding:"required"`
	ManagerId  string `json:"managerId"`
	SourceName string `json:"sourceName" binding:"required"`
}

func GetGoogleOauth(c *gin.Context) {
	var params GoogleOauthParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	claimsObj, ok := c.Get("project-claims")
	if !ok {
		c.JSON(400, gin.H{"error": "missing claims"})
		return
	}
	claims, ok := claimsObj.(*tokens.ProjectToken)
	if !ok {
		c.JSON(400, gin.H{"error": "invalid claims"})
		return
	}
	state, _ := nanoid.Nanoid(5)
	_, err := db.Client.GoogleOauthState.Create().
		SetID(state).
		SetCustomerID(params.CustomerId).
		SetManagerID(params.ManagerId).
		SetSourceName(params.SourceName).
		Save(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create state", "message": err.Error()})
		return
	}
	urlParams := url.Values{}
	urlParams.Add("client_id", os.Getenv("GA_CLIENT_ID"))
	urlParams.Add("redirect_uri", os.Getenv("GA_REDIRECT_URI"))
	urlParams.Add("response_type", "code")
	urlParams.Add("scope", "https://www.googleapis.com/auth/adwords")
	urlParams.Add("access_type", "offline")
	urlParams.Add("prompt", "consent")
	urlParams.Add("state", claims.ProjectId+state)
	c.JSON(200, gin.H{"status": "ok", "url": "https://accounts.google.com/o/oauth2/v2/auth?" + urlParams.Encode()})
}
