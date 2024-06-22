package sipapi

import (
	"fmt"
	"net/url"

	"github.com/panjjo/gosip/utils"
	"github.com/sirupsen/logrus"
)

type zlmGetMediaListReq struct {
	vhost    string
	schema   string
	streamID string
	app      string
}
type zlmGetMediaListResp struct {
	Code int                       `json:"code"`
	Data []zlmGetMediaListDataResp `json:"data"`
}
type zlmGetMediaListDataResp struct {
	App        string                  `json:"app"`
	Stream     string                  `json:"stream"`
	Schema     string                  `json:"schema"`
	OriginType int                     `json:"originType"`
	Tracks     []zlmGetMediaListTracks `json:"tracks"`
}
type zlmGetMediaListTracks struct {
	Type    int `json:"codec_type"`
	CodecID int `json:"codec_id"`
	Height  int `json:"height"`
	Width   int `json:"width"`
	FPS     int `json:"fps"`
}

// zlm 获取流列表信息
func zlmGetMediaList(req zlmGetMediaListReq) zlmGetMediaListResp {
	res := zlmGetMediaListResp{}
	ssrc := req.streamID
	if d, ok := StreamList.Response.Load(ssrc); ok {
		params := d.(*Streams)

		reqStr := "/index/api/getMediaList?secret=" + params.Secret
		if req.streamID != "" {
			reqStr += "&stream=" + req.streamID
		}
		if req.app != "" {
			reqStr += "&app=" + req.app
		}
		if req.schema != "" {
			reqStr += "&schema=" + req.schema
		}
		if req.vhost != "" {
			reqStr += "&vhost=" + req.vhost
		}
		body, err := utils.GetRequest(params.RESTFUL + reqStr)
		if err != nil {
			logrus.Errorln("get stream mediaList fail,", err)
			return res
		}
		if err = utils.JSONDecode(body, &res); err != nil {
			logrus.Errorln("get stream mediaList fail,", err)
			return res
		}
		logrus.Traceln("zlmGetMediaList ", string(body), req.streamID)
	}

	return res
}

var zlmDeviceVFMap = map[int]string{
	0: "H264",
	1: "H265",
	2: "ACC",
	3: "G711A",
	4: "G711U",
}

func transZLMDeviceVF(t int) string {
	if v, ok := zlmDeviceVFMap[t]; ok {
		return v
	}
	return "undefind"
}

type rtpInfo struct {
	Code  int  `json:"code"`
	Exist bool `json:"exist"`
}

// 获取流在zlm上的信息
func zlmGetMediaInfo(ssrc string) rtpInfo {
	res := rtpInfo{}
	if d, ok := StreamList.Response.Load(ssrc); ok {
		params := d.(*Streams)

		body, err := utils.GetRequest(params.RESTFUL + "/index/api/getRtpInfo?secret=" + params.Secret + "&stream_id=" + ssrc)
		if err != nil {
			logrus.Errorln("get stream rtpInfo fail,", err)
			return res
		}
		if err = utils.JSONDecode(body, &res); err != nil {
			logrus.Errorln("get stream rtpInfo fail,", err)
			return res
		}
	}
	return res
}

// zlm 关闭流
func zlmCloseStream(ssrc string) error {
	if d, ok := StreamList.Response.Load(ssrc); ok {
		params := d.(*Streams)

		_, err := utils.GetRequest(params.RESTFUL + "/index/api/close_streams?secret=" + params.Secret + "&stream=" + ssrc)
		if err != nil {
			return err
		}
	}

	return nil
}

// zlm 开始录制视频流
func zlmStartRecord(ssrc string, values url.Values) error {
	if d, ok := StreamList.Response.Load(ssrc); ok {
		params := d.(*Streams)
		body, err := utils.GetRequest(params.RESTFUL + "/index/api/startRecord?" + values.Encode())
		if err != nil {
			return err
		}
		tmp := map[string]interface{}{}
		err = utils.JSONDecode(body, &tmp)
		if err != nil {
			return err
		}
		if code, ok := tmp["code"]; !ok || fmt.Sprint(code) != "0" {
			return utils.NewError(nil, tmp)
		}
	}
	return nil
}

// zlm 停止录制
func zlmStopRecord(ssrc string, values url.Values) error {
	if d, ok := StreamList.Response.Load(ssrc); ok {
		params := d.(*Streams)
		body, err := utils.GetRequest(params.RESTFUL + "/index/api/stopRecord?" + values.Encode())
		if err != nil {
			return err
		}
		tmp := map[string]interface{}{}
		err = utils.JSONDecode(body, &tmp)
		if err != nil {
			return err
		}
		if code, ok := tmp["code"]; !ok || fmt.Sprint(code) != "0" {
			return utils.NewError(nil, tmp)
		}
	}
	return nil
}
