package zkp

const DefaultAgeSchema = `{
  "schema_id": "age_over_18",
  "version": "1.0.0",
  "fields": [
    {"name": "birth_year", "type": "integer", "required": true, "secret": true},
    {"name": "birth_month", "type": "integer", "required": true, "secret": true},
    {"name": "birth_day", "type": "integer", "required": true, "secret": true}
  ],
  "constraints": [
    {"type": "age_verification", "fields": ["birth_year", "birth_month", "birth_day"], "operator": "ge", "value": 18},
    {"type": "range_check", "fields": ["birth_month"], "operator": "between", "value": [1, 12]},
    {"type": "range_check", "fields": ["birth_day"], "operator": "between", "value": [1, 31]}
  ]
}`
