{{/*
Expand the name of the chart.
*/}}
{{- define "kubejanitor-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kubejanitor-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kubejanitor-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kubejanitor-operator.labels" -}}
helm.sh/chart: {{ include "kubejanitor-operator.chart" . }}
{{ include "kubejanitor-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.labels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kubejanitor-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kubejanitor-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kubejanitor-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kubejanitor-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the image name
*/}}
{{- define "kubejanitor-operator.image" -}}
{{- $registry := .Values.global.imageRegistry | default .Values.image.registry -}}
{{- $repository := .Values.image.repository -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- else -}}
{{- printf "%s:%s" $repository $tag -}}
{{- end -}}
{{- end }}

{{/*
Get the namespace for resources
*/}}
{{- define "kubejanitor-operator.namespace" -}}
{{- default .Release.Namespace .Values.namespace -}}
{{- end }}

{{/*
Create the name of the leader election namespace
*/}}
{{- define "kubejanitor-operator.leaderElectionNamespace" -}}
{{- if .Values.manager.leaderElectionNamespace }}
{{- .Values.manager.leaderElectionNamespace }}
{{- else }}
{{- include "kubejanitor-operator.namespace" . }}
{{- end }}
{{- end }}

{{/*
Create the webhook certificate name
*/}}
{{- define "kubejanitor-operator.webhookCertName" -}}
{{- printf "%s-webhook-cert" (include "kubejanitor-operator.fullname" .) }}
{{- end }}

{{/*
Create the webhook service name
*/}}
{{- define "kubejanitor-operator.webhookServiceName" -}}
{{- printf "%s-webhook-service" (include "kubejanitor-operator.fullname" .) }}
{{- end }}

{{/*
Create pod annotations
*/}}
{{- define "kubejanitor-operator.podAnnotations" -}}
{{- with .Values.annotations }}
{{ toYaml . }}
{{- end }}
{{- with .Values.podAnnotations }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Create pod labels
*/}}
{{- define "kubejanitor-operator.podLabels" -}}
{{ include "kubejanitor-operator.selectorLabels" . }}
{{- with .Values.labels }}
{{ toYaml . }}
{{- end }}
{{- with .Values.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}