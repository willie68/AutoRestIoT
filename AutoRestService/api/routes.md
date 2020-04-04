# 



## Routes

<details>
<summary>`/*/api/v1/config/*`</summary>

- [SetContentType.func1]()
- [Logger]()
- [DefaultCompress]()
- [Recoverer]()
- [(*SysAPIKey).Handler-fm]()
- **/***
	- **/api/v1/config/***
		- **/**
			- _GET_
				- [GetConfigEndpoint]()
			- _DELETE_
				- [DeleteConfigEndpoint]()
			- _POST_
				- [PostConfigEndpoint]()

</details>
<details>
<summary>`/*/api/v1/config/*/size`</summary>

- [SetContentType.func1]()
- [Logger]()
- [DefaultCompress]()
- [Recoverer]()
- [(*SysAPIKey).Handler-fm]()
- **/***
	- **/api/v1/config/***
		- **/size**
			- _GET_
				- [GetConfigSizeEndpoint]()

</details>
<details>
<summary>`/*/health/*/health`</summary>

- [SetContentType.func1]()
- [Logger]()
- [DefaultCompress]()
- [Recoverer]()
- [(*SysAPIKey).Handler-fm]()
- **/***
	- **/health/***
		- **/health**
			- _GET_
				- [GetHealthyEndpoint]()

</details>
<details>
<summary>`/*/health/*/readiness`</summary>

- [SetContentType.func1]()
- [Logger]()
- [DefaultCompress]()
- [Recoverer]()
- [(*SysAPIKey).Handler-fm]()
- **/***
	- **/health/***
		- **/readiness**
			- _GET_
				- [GetReadinessEndpoint]()

</details>

Total # of routes: 4