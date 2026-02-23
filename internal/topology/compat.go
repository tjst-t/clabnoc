package topology

// NormalizeLinks converts links from either v0.73+ or legacy format
// into a consistent internal representation. This is handled during
// parsing in convertLink - this file documents the compatibility approach.
//
// v0.73.0+ format:
//
//	{
//	  "endpoints": {
//	    "a": { "node": "...", "interface": "...", "mac": "...", "peer": "z" },
//	    "z": { "node": "...", "interface": "...", "mac": "...", "peer": "a" }
//	  }
//	}
//
// Legacy format:
//
//	{
//	  "a": { "node": "...", "interface": "...", "mac": "..." },
//	  "z": { "node": "...", "interface": "...", "mac": "..." }
//	}
//
// Both formats are unified into the Link type with separate A/Z endpoints.
