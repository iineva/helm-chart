{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "externalServices.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "externalServices.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "externalServices.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Generate metadata
*/}}
{{- define "externalService.metadata" -}}
{{- $fullName := "" -}}
{{- with .spec.fullName -}}
{{- $fullName = . -}}
{{- else -}}
{{- $fullName = (printf "%s-%s" .Release.Name .name) -}}
{{- end -}}
metadata:
  name: {{ $fullName }}
  namespace: {{ .spec.namespace | default .Release.Namespace }}
  labels:
    app: {{ $fullName }}
    chart: {{ include "externalServices.chart" . }}
    release: {{ .Release.Name }}
    heritage: {{ .Release.Service }}
{{- end -}}
