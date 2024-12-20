# MIT License
#
# (C) Copyright [2022-2023] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

FROM alpine/git AS build

RUN git clone --branch v5.1.0 --single-branch https://github.com/Cray-HPE/hms-test.git

FROM docker.io/library/python:3.13-slim

RUN pip install \
    pytest==7.1.2 \
    tavern==1.23.1 \
    allure-pytest==2.12.0 \
    setuptools


# Copy over what we need from hms-test

COPY --from=build /git/hms-test/cmd/hms-pytest /usr/bin/hms-pytest
COPY --from=build /git/hms-test/cmd/entrypoint.sh /usr/bin/entrypoint.sh
COPY --from=build /git/hms-test/cmd/smoke_test.py /src/app/smoke_test.py
COPY --from=build /git/hms-test/cmd/smoke_pytest /src/app/smoke_pytest
COPY --from=build /git/hms-test/cmd/tavern_test.py /src/app/tavern_test.py
COPY --from=build /git/hms-test/libs/pytest.ini /src/app/pytest.ini
COPY --from=build /git/hms-test/libs/ /src/libs

COPY smoke/ /src/app/
COPY api/ /src/app/api
COPY tavern_global_config_ct_test_emulated_hardware.yaml /src/app/tavern_global_config_ct_test_emulated_hardware.yaml
COPY tavern_global_config_ct_test_production.yaml /src/app/tavern_global_config_ct_test_production.yaml

ENV PATH="/src/libs:${PATH}"
ENV PATH="/src/app:${PATH}"

