# Fill Q Challenge API Digestion

Repository code is used in AWS Lambda to digest different APIS and return one simplified object.

## Description

Firstly, code requests information for each player from Riot Games API and twitch API. After the request is completed, the returned objects are joined. The merged object is trimmed down to the information that is needed for the website and returned per AWS Gateway REST request.
