import React, { useEffect } from 'react'
import { useQuery, gql, useMutation } from '@apollo/client'
import { INITIAL_SETUP_QUERY, login } from './loginUtilities'

import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { CheckInitialSetup } from './__generated__/CheckInitialSetup'
import { Authorize, AuthorizeVariables } from './__generated__/Authorize'

const authorizeMutation = gql`
  mutation Authorize($username: String!, $password: String!) {
    authorizeUser(username: $username, password: $password) {
      success
      status
      token
    }
  }
`

const LoginPage = () => {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [authorize] = useMutation<
    Authorize,
    AuthorizeVariables
  >(authorizeMutation)

  const { data: initialSetupData } = useQuery<CheckInitialSetup>(
    INITIAL_SETUP_QUERY,
    { variables: {},
    onCompleted: data => {
      if (data?.siteInfo?.initialSetup) {
        navigate('/initialSetup')
      }
      else {
        authorize({
            variables: {
              username: "admin",
              password: "admin",
            },
            onCompleted: data => {
              const {success, token} = data.authorizeUser

              if (success && token) {
                login(token)
              }
            }
          }
        )
      }
    }}
  )

  return null
}

export default LoginPage
