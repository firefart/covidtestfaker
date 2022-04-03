#!/bin/sh

echo "Copying unit file"
cp /home/covidtestfaker/covidtestfaker.service /etc/systemd/system/covidtestfaker.service
echo "reloading systemctl"
systemctl daemon-reload
echo "enabling service"
systemctl enable covidtestfaker.service
systemctl start covidtestfaker.service
systemctl status covidtestfaker.service
