#!/bin/bash
gcloud functions deploy ImportQrz \
  --project k0swe-kellog \
  --entry-point ImportQrz \
  --runtime go113\
  --trigger-http

gcloud functions deploy ImportLotw \
  --project k0swe-kellog \
  --entry-point ImportLotw \
  --runtime go113\
  --trigger-http
