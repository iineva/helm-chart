{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "api-server.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "api-server.fullname" -}}
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
{{- define "api-server.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "api-server.labels" -}}
helm.sh/chart: {{ include "api-server.chart" . }}
{{ include "api-server.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "api-server.selectorLabels" -}}
{{- if .name }}
app.kubernetes.io/name: {{ include "api-server.name" . }}-{{ .name }}
app.kubernetes.io/instance: {{ .Release.Name }}-{{ .name }}
{{- else -}}
app.kubernetes.io/name: {{ include "api-server.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "api-server.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "api-server.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{- define "api-server.sortAlphaKeyValue" -}}
{{- range $key := (keys .) | sortAlpha -}}
{{ $key }}: "{{ get $ $key }}"
{{- end -}}
{{- end -}}

{{- define "api-server.configHash" -}}
{{- range $name, $configs := .root.Values.configMaps }}
{{- if eq $name (default "" $.name) -}}
{{- include "api-server.sortAlphaKeyValue" $configs | sha256sum -}}
{{- end -}}
{{- end -}}
{{- end -}}