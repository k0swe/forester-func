#!/bin/bash
gcloud functions deploy ImportQrz \
  --project k0swe-kellog \
  --entry-point ImportQrz \
  --runtime go113\
  --trigger-http
