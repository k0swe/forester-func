#!/bin/bash
gcloud functions deploy HelloHTTP \
  --project k0swe-kellog \
  --entry-point HelloHTTP \
  --runtime go113\
  --trigger-http
